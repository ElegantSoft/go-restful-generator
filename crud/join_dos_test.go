package crud

// Regression tests for ClickUp 1247n713gnj (DOS-01): "Nested Join OOM —
// unauthenticated single request crashes the backend".
//
// The `join` query param used to be handed straight to GORM's Preload with no
// depth limit and no validation, so a request such as
//
//	GET /products?join=Categories.Products.Categories.Products.Categories.Products.Categories.Products.Categories&limit=10
//
// made GORM issue one query per hop and build an exponentially growing object
// graph, OOM-killing the process.
//
// The models below mirror the shape from the ticket: Product <-> Category is a
// many2many cycle (the amplification vector) and Product has many Variants.
//
// Test groups:
//   - TestJoinDoS_*     reproduce the attack. They FAIL on the vulnerable code
//                       (the request is accepted / hits the DB) and PASS once
//                       the fix rejects it up front, before any SQL is executed.
//   - TestJoinCompat_*  legitimate joins services rely on. They must pass both
//                       BEFORE and AFTER the fix (no behavior change).

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dosProduct struct {
	ID         uint `gorm:"primaryKey"`
	Name       string
	Categories []dosCategory `gorm:"many2many:dos_product_categories"`
	Variants   []dosVariant  `gorm:"foreignKey:ProductID"`
}

type dosCategory struct {
	ID       uint `gorm:"primaryKey"`
	Name     string
	Products []dosProduct `gorm:"many2many:dos_product_categories"`
}

type dosVariant struct {
	ID        uint `gorm:"primaryKey"`
	ProductID uint
	Name      string
	Product   *dosProduct `gorm:"foreignKey:ProductID"`
}

const (
	// exact payload from the ticket's proof of concept (9 hops)
	attackJoin9 = "Categories.Products.Categories.Products.Categories.Products.Categories.Products.Categories"
	// the slower 11-hop variant from the ticket
	attackJoin11 = attackJoin9 + ".Products.Categories"

	dosProductsCount   = 3
	dosCategoriesCount = 2
)

type dosEnv struct {
	svc     *Service[dosProduct]
	db      *gorm.DB
	queries int64
}

// queryCount is the number of SELECTs (root, count and every preload hop)
// executed since the last reset.
func (e *dosEnv) queryCount() int64 { return atomic.LoadInt64(&e.queries) }
func (e *dosEnv) resetQueries()     { atomic.StoreInt64(&e.queries, 0) }

// newDosEnv creates an isolated in-memory SQLite DB where every product belongs
// to every category, so each Categories.Products hop multiplies the graph.
func newDosEnv(t *testing.T) *dosEnv {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := gdb.AutoMigrate(&dosCategory{}, &dosProduct{}, &dosVariant{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	cats := make([]dosCategory, dosCategoriesCount)
	for i := range cats {
		cats[i] = dosCategory{Name: fmt.Sprintf("cat-%d", i+1)}
	}
	if err := gdb.Create(&cats).Error; err != nil {
		t.Fatalf("seed categories: %v", err)
	}
	for i := 0; i < dosProductsCount; i++ {
		p := dosProduct{
			Name:       fmt.Sprintf("product-%d", i+1),
			Categories: cats,
			Variants:   []dosVariant{{Name: fmt.Sprintf("variant-%d", i+1)}},
		}
		if err := gdb.Create(&p).Error; err != nil {
			t.Fatalf("seed product: %v", err)
		}
	}

	env := &dosEnv{db: gdb}
	if err := gdb.Callback().Query().Before("gorm:query").Register("test:count_queries", func(*gorm.DB) {
		atomic.AddInt64(&env.queries, 1)
	}); err != nil {
		t.Fatalf("register query counter: %v", err)
	}
	env.svc = NewService[dosProduct](NewRepository[dosProduct](gdb, dosProduct{}))
	env.resetQueries()
	return env
}

// dosRouter mirrors the generated controller's findAll handler: bind the query,
// call Service.Find and map any error to HTTP 400.
func dosRouter(svc *Service[dosProduct]) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/products", func(ctx *gin.Context) {
		var api GetAllRequest
		if err := ctx.ShouldBindQuery(&api); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		var result []dosProduct
		var totalRows int64
		if err := svc.Find(api, &result, &totalRows); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	})
	return r
}

func jsonSize(v interface{}) int {
	b, _ := json.Marshal(v)
	return len(b)
}

// ---------------------------------------------------------------------------
// Attack reproduction (fail before the fix, pass after it)
// ---------------------------------------------------------------------------

// maliciousJoins are join values that must be refused without touching the DB.
var maliciousJoins = []struct {
	name string
	join string
}{
	{"ticket_poc_9_hops", attackJoin9},
	{"ticket_11_hops", attackJoin11},
	{"lowercase_9_hops", strings.ToLower(attackJoin9)},
	{"padded_with_spaces", "  " + attackJoin9 + "  "},
	{"valid_join_then_attack", "Variants," + attackJoin9},
	{"attack_plus_default_join", attackJoin9 + ",categories"}, // posts controller appends ",category"
	{"four_levels_over_default_cap", "Categories.Products.Categories.Products"},
	{"repeated_relation_within_cap", "Categories.Products.Categories"},
	{"wildcard_nesting", "*.*.*.*.*.*.*.*.*"},
	{"wildcard_single", "*"},
	{"unknown_relation", "Bogus"},
	{"unknown_nested_relation", "Categories.Bogus"},
	{"empty_segment", "Categories..Products"},
}

func TestJoinDoS_FindRejectsMaliciousJoinBeforeAnyQuery(t *testing.T) {
	env := newDosEnv(t)

	for _, tc := range maliciousJoins {
		t.Run(tc.name, func(t *testing.T) {
			env.resetQueries()
			var out []dosProduct
			var total int64
			err := env.svc.Find(GetAllRequest{Join: tc.join, Limit: 10}, &out, &total)
			if err == nil {
				t.Fatalf("VULNERABLE: join %q was accepted: %d SQL queries executed, response would be %d bytes",
					tc.join, env.queryCount(), jsonSize(out))
			}
			if n := env.queryCount(); n != 0 {
				t.Fatalf("join %q was rejected only after %d SQL queries (err=%v); it must be rejected before touching the DB",
					tc.join, n, err)
			}
		})
	}
}

func TestJoinDoS_FindOneRejectsMaliciousJoinBeforeAnyQuery(t *testing.T) {
	env := newDosEnv(t)

	for _, tc := range maliciousJoins {
		t.Run(tc.name, func(t *testing.T) {
			env.resetQueries()
			var out dosProduct
			err := env.svc.FindOne(GetAllRequest{Join: tc.join, Filter: []string{"id||eq||1"}}, &out)
			if err == nil {
				t.Fatalf("VULNERABLE: join %q was accepted by FindOne: %d SQL queries executed, response would be %d bytes",
					tc.join, env.queryCount(), jsonSize(out))
			}
			if n := env.queryCount(); n != 0 {
				t.Fatalf("join %q was rejected only after %d SQL queries (err=%v); it must be rejected before touching the DB",
					tc.join, n, err)
			}
		})
	}
}

// TestJoinDoS_HTTPAttackReturns400 replays the exact request from the ticket
// (plus the Redis cache-bypass variant) through a gin handler identical to the
// generated controllers and expects an immediate 400 with no DB work.
func TestJoinDoS_HTTPAttackReturns400(t *testing.T) {
	env := newDosEnv(t)
	router := dosRouter(env.svc)

	// baseline: size of a normal 1-level join response, to show amplification
	baseline := httptest.NewRecorder()
	router.ServeHTTP(baseline, httptest.NewRequest(http.MethodGet, "/products?join=Categories&limit=10", nil))
	if baseline.Code != http.StatusOK {
		t.Fatalf("baseline join=Categories: status %d, body %s", baseline.Code, baseline.Body.String())
	}

	targets := []string{
		"/products?join=" + url.QueryEscape(attackJoin9) + "&limit=10",
		"/products?join=" + url.QueryEscape(attackJoin9) + "&limit=10&_=1", // cache bypass
		"/products?join=" + url.QueryEscape(attackJoin11) + "&limit=10&_=2",
	}
	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			env.resetQueries()
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("VULNERABLE: status %d, %d SQL queries, %d-byte body (a normal 1-level join is %d bytes => x%d amplification)",
					rec.Code, env.queryCount(), rec.Body.Len(), baseline.Body.Len(), rec.Body.Len()/baseline.Body.Len())
			}
			if n := env.queryCount(); n != 0 {
				t.Fatalf("attack answered 400 only after %d SQL queries; must be rejected before touching the DB", n)
			}
			if !strings.Contains(strings.ToLower(rec.Body.String()), "join") {
				t.Fatalf("400 body should explain the join problem, got %s", rec.Body.String())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Backward compatibility (must pass before AND after the fix)
// ---------------------------------------------------------------------------

func TestJoinCompat_LegitJoinsStillWork(t *testing.T) {
	env := newDosEnv(t)

	find := func(t *testing.T, join string) []dosProduct {
		t.Helper()
		var out []dosProduct
		var total int64
		if err := env.svc.Find(GetAllRequest{Join: join, Limit: 10}, &out, &total); err != nil {
			t.Fatalf("join %q: unexpected error: %v", join, err)
		}
		if total != dosProductsCount || len(out) != dosProductsCount {
			t.Fatalf("join %q: total=%d len=%d, want %d", join, total, len(out), dosProductsCount)
		}
		return out
	}

	t.Run("no_join", func(t *testing.T) {
		for _, p := range find(t, "") {
			if p.Categories != nil || p.Variants != nil {
				t.Fatalf("nothing should be preloaded without join: %+v", p)
			}
		}
	})

	for _, join := range []string{"Categories", "categories", ",categories", "Categories,"} {
		join := join
		t.Run("single_"+join, func(t *testing.T) {
			for _, p := range find(t, join) {
				if len(p.Categories) != dosCategoriesCount {
					t.Fatalf("join %q: product %d has %d categories, want %d", join, p.ID, len(p.Categories), dosCategoriesCount)
				}
				if p.Variants != nil {
					t.Fatalf("join %q: variants must not be preloaded", join)
				}
				for _, c := range p.Categories {
					if c.Products != nil {
						t.Fatalf("join %q: nested products must not be preloaded", join)
					}
				}
			}
		})
	}

	t.Run("multiple_relations", func(t *testing.T) {
		for _, p := range find(t, "Categories,Variants") {
			if len(p.Categories) != dosCategoriesCount || len(p.Variants) != 1 {
				t.Fatalf("product %d: categories=%d variants=%d", p.ID, len(p.Categories), len(p.Variants))
			}
		}
	})

	t.Run("nested_two_levels", func(t *testing.T) {
		for _, join := range []string{"Categories.Products", "categories.products"} {
			for _, p := range find(t, join) {
				if len(p.Categories) != dosCategoriesCount {
					t.Fatalf("join %q: product %d has %d categories", join, p.ID, len(p.Categories))
				}
				for _, c := range p.Categories {
					if len(c.Products) != dosProductsCount {
						t.Fatalf("join %q: category %d has %d products, want %d", join, c.ID, len(c.Products), dosProductsCount)
					}
				}
			}
		}
	})

	t.Run("nested_three_levels", func(t *testing.T) {
		for _, join := range []string{"Categories.Products.Variants", "categories.products.variants"} {
			for _, p := range find(t, join) {
				if len(p.Categories) != dosCategoriesCount {
					t.Fatalf("join %q: product %d has %d categories", join, p.ID, len(p.Categories))
				}
				for _, c := range p.Categories {
					if len(c.Products) != dosProductsCount {
						t.Fatalf("join %q: category %d has %d products, want %d", join, c.ID, len(c.Products), dosProductsCount)
					}
					for _, cp := range c.Products {
						if len(cp.Variants) != 1 {
							t.Fatalf("join %q: nested product %d has %d variants, want 1", join, cp.ID, len(cp.Variants))
						}
					}
				}
			}
		}
	})

	t.Run("parent_and_nested_together", func(t *testing.T) {
		out := find(t, "Categories,Categories.Products,Variants")
		if len(out[0].Categories) != dosCategoriesCount || len(out[0].Categories[0].Products) != dosProductsCount || len(out[0].Variants) != 1 {
			t.Fatalf("unexpected graph: %+v", out[0])
		}
	})

	t.Run("pagination_with_join", func(t *testing.T) {
		var out []dosProduct
		var total int64
		if err := env.svc.Find(GetAllRequest{Join: "Categories", Limit: 2, Page: 2}, &out, &total); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != dosProductsCount || len(out) != 1 || len(out[0].Categories) != dosCategoriesCount {
			t.Fatalf("total=%d len=%d", total, len(out))
		}
	})

	t.Run("find_one_with_nested_join", func(t *testing.T) {
		var out dosProduct
		if err := env.svc.FindOne(GetAllRequest{Join: "Categories.Products,variants", Filter: []string{"id||eq||2"}}, &out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.ID != 2 || len(out.Categories) != dosCategoriesCount || len(out.Categories[0].Products) != dosProductsCount || len(out.Variants) != 1 {
			t.Fatalf("unexpected result: %+v", out)
		}
	})

	t.Run("http_legit_join_ok", func(t *testing.T) {
		rec := httptest.NewRecorder()
		dosRouter(env.svc).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/products?join=Categories.Products,Variants&limit=10", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d, body %s", rec.Code, rec.Body.String())
		}
	})
}
