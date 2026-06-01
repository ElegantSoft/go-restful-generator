package crud

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test-local models mirror db/models.Post & Category (same columns and the
// Category relation) but drop the Postgres-only `uuid_generate_v4()` default so
// they migrate cleanly on SQLite. The security fixes validate against whatever
// model the service is built with, so these exercise the exact same code paths.
type testCategory struct {
	ID        uuid.UUID `gorm:"type:text;primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type testPost struct {
	ID          uuid.UUID `gorm:"type:text;primaryKey"`
	Title       string
	Description string
	CategoryID  uuid.NullUUID `gorm:"type:text"`
	Category    *testCategory
	Price       uint32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type seedData struct {
	catTech uuid.UUID
	catFood uuid.UUID
	byTitle map[string]uuid.UUID
}

// newTestService spins up a fresh in-memory SQLite DB, migrates the schema, seeds
// a deterministic dataset, and returns a Service plus the seeded identifiers.
func newTestService(t *testing.T) (*Service[testPost], *gorm.DB, seedData) {
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
	sqlDB.SetMaxOpenConns(1) // keep the shared in-memory DB alive for the whole test
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := gdb.AutoMigrate(&testCategory{}, &testPost{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	seed := seedData{
		catTech: uuid.New(),
		catFood: uuid.New(),
		byTitle: map[string]uuid.UUID{},
	}
	cats := []testCategory{
		{ID: seed.catTech, Name: "Tech"},
		{ID: seed.catFood, Name: "Food"},
	}
	if err := gdb.Create(&cats).Error; err != nil {
		t.Fatalf("seed categories: %v", err)
	}

	tech := uuid.NullUUID{UUID: seed.catTech, Valid: true}
	food := uuid.NullUUID{UUID: seed.catFood, Valid: true}
	none := uuid.NullUUID{}

	// 8 posts: prices spread for ranges/sort; 2 null-category rows; 2 "Go" titles.
	posts := []testPost{
		{Title: "Go in Action", Price: 100, CategoryID: tech},
		{Title: "Go Web Dev", Price: 150, CategoryID: tech},
		{Title: "Rust Book", Price: 200, CategoryID: tech},
		{Title: "Cooking 101", Price: 50, CategoryID: food},
		{Title: "Baking Bread", Price: 75, CategoryID: food},
		{Title: "Sale Items", Price: 500, CategoryID: food},
		{Title: "Uncategorized A", Price: 300, CategoryID: none},
		{Title: "Uncategorized B", Price: 250, CategoryID: none},
	}
	for i := range posts {
		posts[i].ID = uuid.New()
		seed.byTitle[posts[i].Title] = posts[i].ID
	}
	if err := gdb.Create(&posts).Error; err != nil {
		t.Fatalf("seed posts: %v", err)
	}

	svc := NewService[testPost](NewRepository[testPost](gdb, testPost{}))
	return svc, gdb, seed
}

func findAll(t *testing.T, svc *Service[testPost], api GetAllRequest) ([]testPost, int64) {
	t.Helper()
	if api.Limit == 0 {
		api.Limit = 100
	}
	var out []testPost
	var total int64
	if err := svc.Find(api, &out, &total); err != nil {
		t.Fatalf("Find(%+v): %v", api, err)
	}
	return out, total
}

func prices(posts []testPost) []uint32 {
	out := make([]uint32, len(posts))
	for i, p := range posts {
		out[i] = p.Price
	}
	return out
}

// ---------------------------------------------------------------------------
// Functional / business tests (core features must keep working)
// ---------------------------------------------------------------------------

func TestPagination(t *testing.T) {
	svc, _, _ := newTestService(t)

	page1, total := findAll(t, svc, GetAllRequest{Limit: 3, Page: 1})
	if total != 8 {
		t.Fatalf("total = %d, want 8", total)
	}
	if len(page1) != 3 {
		t.Fatalf("page1 len = %d, want 3", len(page1))
	}

	page2, _ := findAll(t, svc, GetAllRequest{Limit: 3, Page: 2})
	if len(page2) != 3 {
		t.Fatalf("page2 len = %d, want 3", len(page2))
	}
	// pages must not overlap
	seen := map[uuid.UUID]bool{}
	for _, p := range page1 {
		seen[p.ID] = true
	}
	for _, p := range page2 {
		if seen[p.ID] {
			t.Fatalf("page2 overlaps page1 on %s", p.ID)
		}
	}
}

func TestFieldSelection(t *testing.T) {
	svc, _, _ := newTestService(t)

	out, _ := findAll(t, svc, GetAllRequest{Fields: "id,title"})
	if len(out) != 8 {
		t.Fatalf("len = %d, want 8", len(out))
	}
	for _, p := range out {
		if p.Title == "" {
			t.Fatalf("title not selected: %+v", p)
		}
		if p.Price != 0 {
			t.Fatalf("price should be unselected (zero), got %d", p.Price)
		}
	}
}

func TestFieldSelectionTrimsAndDropsInvalid(t *testing.T) {
	svc, _, _ := newTestService(t)
	// " title " trimmed & kept; "bogus" dropped; query still succeeds.
	out, _ := findAll(t, svc, GetAllRequest{Fields: " title , bogus"})
	if len(out) != 8 {
		t.Fatalf("len = %d, want 8", len(out))
	}
	for _, p := range out {
		if p.Title == "" {
			t.Fatalf("title should be selected")
		}
	}
}

// TestNamingConventionEquivalence proves that resolveColumn accepts BOTH the DB
// column name (snake_case, e.g. category_id) and the Go struct field name
// (PascalCase, e.g. CategoryID), normalizing both to the same canonical column.
// It also documents that unrelated casings (camelCase / SHOUTING) are rejected.
func TestNamingConventionEquivalence(t *testing.T) {
	svc, _, _ := newTestService(t)

	t.Run("filter_db_vs_struct_name", func(t *testing.T) {
		snake, _ := findAll(t, svc, GetAllRequest{Filter: []string{"price||eq||100"}})
		pascal, _ := findAll(t, svc, GetAllRequest{Filter: []string{"Price||eq||100"}})
		if len(snake) != 1 || len(pascal) != 1 {
			t.Fatalf("price=%d Price=%d, want 1/1", len(snake), len(pascal))
		}
		if snake[0].ID != pascal[0].ID {
			t.Fatalf("price vs Price returned different rows")
		}
	})

	t.Run("filter_multiword_column", func(t *testing.T) {
		// category_id (DB) and CategoryID (struct) must behave identically.
		bySnake, _ := findAll(t, svc, GetAllRequest{Filter: []string{"category_id||isnull"}})
		byPascal, _ := findAll(t, svc, GetAllRequest{Filter: []string{"CategoryID||isnull"}})
		if len(bySnake) != 2 || len(byPascal) != 2 {
			t.Fatalf("category_id=%d CategoryID=%d, want 2/2", len(bySnake), len(byPascal))
		}
	})

	t.Run("fields_db_vs_struct_name", func(t *testing.T) {
		snake, _ := findAll(t, svc, GetAllRequest{Fields: "id,category_id"})
		pascal, _ := findAll(t, svc, GetAllRequest{Fields: "ID,CategoryID"})
		if len(snake) != 8 || len(pascal) != 8 {
			t.Fatalf("len snake=%d pascal=%d, want 8/8", len(snake), len(pascal))
		}
		// Both forms must select the same column: title was NOT selected -> empty.
		for i := range snake {
			if snake[i].Title != "" || pascal[i].Title != "" {
				t.Fatalf("title should be unselected in both naming forms")
			}
		}
	})

	t.Run("sort_struct_name", func(t *testing.T) {
		// CreatedAt (struct) resolves to created_at; should not be dropped.
		out, _ := findAll(t, svc, GetAllRequest{Sort: []string{"CreatedAt,asc"}})
		if len(out) != 8 {
			t.Fatalf("sort by CreatedAt returned %d rows, want 8", len(out))
		}
	})

	t.Run("unrelated_casing_rejected", func(t *testing.T) {
		// Neither the DB name nor the exact Go field name -> dropped, so the
		// legit filter alone decides the result (1 row), not widened to all.
		out, _ := findAll(t, svc, GetAllRequest{
			Filter: []string{"price||eq||100", "categoryId||isnull", "CATEGORY_ID||isnull"},
		})
		if len(out) != 1 {
			t.Fatalf("unrecognized casings should be dropped; got %d rows, want 1", len(out))
		}
	})
}

func TestFilterOperators(t *testing.T) {
	svc, _, seed := newTestService(t)

	cases := []struct {
		name   string
		filter []string
		want   int
	}{
		{"eq", []string{"price||eq||100"}, 1},
		{"ne", []string{"price||ne||100"}, 7},
		{"gt", []string{"price||gt||200"}, 3},   // 500,300,250
		{"gte", []string{"price||gte||200"}, 4}, // 200,500,300,250
		{"lt", []string{"price||lt||100"}, 2},   // 50,75
		{"lte", []string{"price||lte||100"}, 3}, // 50,75,100
		{"isnull", []string{"category_id||isnull"}, 2},
		{"notnull", []string{"category_id||notnull"}, 6},
		{"and_combo", []string{"price||gte||100", "price||lte||200"}, 3}, // 100,150,200
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _ := findAll(t, svc, GetAllRequest{Filter: tc.filter})
			if len(out) != tc.want {
				t.Fatalf("filter %v => %d rows, want %d (prices=%v)", tc.filter, len(out), tc.want, prices(out))
			}
		})
	}

	t.Run("in", func(t *testing.T) {
		p1 := seed.byTitle["Go in Action"]
		p4 := seed.byTitle["Cooking 101"]
		out, _ := findAll(t, svc, GetAllRequest{
			Filter: []string{fmt.Sprintf("id||$in||%s,%s", p1, p4)},
		})
		if len(out) != 2 {
			t.Fatalf("$in => %d rows, want 2", len(out))
		}
	})
}

func TestSorting(t *testing.T) {
	svc, _, _ := newTestService(t)

	asc, _ := findAll(t, svc, GetAllRequest{Sort: []string{"price,asc"}})
	if asc[0].Price != 50 {
		t.Fatalf("asc first price = %d, want 50 (got %v)", asc[0].Price, prices(asc))
	}

	desc, _ := findAll(t, svc, GetAllRequest{Sort: []string{"price,desc"}})
	if desc[0].Price != 500 {
		t.Fatalf("desc first price = %d, want 500 (got %v)", desc[0].Price, prices(desc))
	}

	def, _ := findAll(t, svc, GetAllRequest{Sort: []string{"price"}})
	if def[0].Price != 500 {
		t.Fatalf("default first price = %d, want 500 (desc default)", def[0].Price)
	}
}

func TestJoinPreload(t *testing.T) {
	svc, _, _ := newTestService(t)

	out, _ := findAll(t, svc, GetAllRequest{
		Join:   "Category",
		Filter: []string{"price||eq||100"}, // "Go in Action" -> Tech
	})
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].Category == nil {
		t.Fatalf("Category not preloaded")
	}
	if out[0].Category.Name != "Tech" {
		t.Fatalf("Category.Name = %q, want Tech", out[0].Category.Name)
	}
}

func TestSearchAndOr(t *testing.T) {
	svc, _, _ := newTestService(t)

	t.Run("and", func(t *testing.T) {
		out, _ := findAll(t, svc, GetAllRequest{
			S: `{"$and":[{"price":{"gte":"100"}},{"price":{"lte":"200"}}]}`,
		})
		if len(out) != 3 {
			t.Fatalf("$and => %d rows, want 3 (prices=%v)", len(out), prices(out))
		}
	})

	t.Run("or", func(t *testing.T) {
		out, _ := findAll(t, svc, GetAllRequest{
			S: `{"$or":[{"price":{"eq":"50"}},{"price":{"eq":"500"}}]}`,
		})
		if len(out) != 2 {
			t.Fatalf("$or => %d rows, want 2 (prices=%v)", len(out), prices(out))
		}
	})

	t.Run("equality_shorthand", func(t *testing.T) {
		out, _ := findAll(t, svc, GetAllRequest{
			S: `{"$and":[{"title":"Rust Book"}]}`,
		})
		if len(out) != 1 {
			t.Fatalf("shorthand => %d rows, want 1", len(out))
		}
		if out[0].Title != "Rust Book" {
			t.Fatalf("title = %q", out[0].Title)
		}
	})
}

// ---------------------------------------------------------------------------
// Security regression tests (injection must be closed, features intact)
// ---------------------------------------------------------------------------

func TestSecurityFieldsSubqueryDropped(t *testing.T) {
	svc, _, _ := newTestService(t)
	// Invalid "column" is a subquery: must be dropped, "id" kept, no error/leak.
	out, total := findAll(t, svc, GetAllRequest{
		Fields: "id,(SELECT name FROM test_categories LIMIT 1)",
	})
	if total != 8 || len(out) != 8 {
		t.Fatalf("got total=%d len=%d, want 8/8", total, len(out))
	}
}

func TestSecuritySortInjectionDropped(t *testing.T) {
	svc, _, _ := newTestService(t)
	out, _ := findAll(t, svc, GetAllRequest{
		Sort: []string{"(CASE WHEN 1=1 THEN id ELSE title END)"},
	})
	if len(out) != 8 {
		t.Fatalf("len = %d, want 8 (injection dropped, query still runs)", len(out))
	}
}

func TestSecurityFilterBooleanBypassDropped(t *testing.T) {
	svc, _, _ := newTestService(t)
	// Legit filter narrows to 1 row; the injected pseudo-column must be dropped,
	// NOT widen the result back to all rows.
	out, _ := findAll(t, svc, GetAllRequest{
		Filter: []string{"price||eq||100", "1) OR (1=1||isnull"},
	})
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1 (boolean bypass must be dropped)", len(out))
	}
}

func TestSecurityInNonStringNoPanic(t *testing.T) {
	svc, _, _ := newTestService(t)
	// Non-string $in value used to panic on value.(string); must be skipped now.
	out, total := findAll(t, svc, GetAllRequest{
		S: `{"$or":[{"price":{"$in":123}}]}`,
	})
	if total != 8 || len(out) != 8 {
		t.Fatalf("got total=%d len=%d, want 8/8 (condition skipped, no panic)", total, len(out))
	}
}

func TestSecurityUnknownColumnInSearchDropped(t *testing.T) {
	svc, _, _ := newTestService(t)
	out, _ := findAll(t, svc, GetAllRequest{
		S: `{"$and":[{"1=1) --":{"eq":"x"}}]}`,
	})
	if len(out) != 8 {
		t.Fatalf("len = %d, want 8 (unknown column dropped)", len(out))
	}
}

// DryRun assertions: the generated SQL must parameterize values and quote the
// resolved column, and must never contain a raw injected identifier verbatim.
func TestSecurityGeneratedSQLIsSafe(t *testing.T) {
	_, gdb, _ := newTestService(t)
	q := &QueryToDBConverter{}

	t.Run("cont_parameterized", func(t *testing.T) {
		tx := gdb.Session(&gorm.Session{DryRun: true}).Model(&testPost{})
		q.filterMapper([]string{"title||cont||go"}, tx)
		var dst []testPost
		tx.Find(&dst)
		sql := tx.Statement.SQL.String()
		if !strings.Contains(sql, "ILIKE ?") {
			t.Fatalf("expected parameterized ILIKE, got: %s", sql)
		}
		if !strings.Contains(sql, "`title`") {
			t.Fatalf("expected quoted column, got: %s", sql)
		}
		foundVar := false
		for _, v := range tx.Statement.Vars {
			if v == "%go%" {
				foundVar = true
			}
		}
		if !foundVar {
			t.Fatalf("expected %%go%% bound as a var, got vars: %v", tx.Statement.Vars)
		}
	})

	t.Run("sort_injection_absent", func(t *testing.T) {
		tx := gdb.Session(&gorm.Session{DryRun: true}).Model(&testPost{})
		q.sortMapper([]string{"(CASE WHEN 1=1 THEN id END)"}, tx)
		var dst []testPost
		tx.Find(&dst)
		sql := tx.Statement.SQL.String()
		if strings.Contains(sql, "CASE WHEN") {
			t.Fatalf("injection payload leaked into SQL: %s", sql)
		}
		if strings.Contains(strings.ToUpper(sql), "ORDER BY") {
			t.Fatalf("dropped sort should not emit ORDER BY: %s", sql)
		}
	})
}
