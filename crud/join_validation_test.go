package crud

import (
	"errors"
	"testing"
)

func TestJoinValidation_SentinelErrors(t *testing.T) {
	env := newDosEnv(t)

	cases := []struct {
		name string
		join string
		want error
	}{
		{"depth_exceeded", "Categories.Products.Categories.Products", ErrJoinDepthExceeded},
		{"unknown_relation", "Nope", ErrInvalidJoin},
		{"empty_segment", "Categories..Products", ErrInvalidJoin},
		{"wildcard", "*", ErrInvalidJoin},
		{"repeated_relation", "Categories.Products.Categories", ErrInvalidJoin},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out []dosProduct
			var total int64
			err := env.svc.Find(GetAllRequest{Join: tc.join}, &out, &total)
			if !errors.Is(err, tc.want) {
				t.Fatalf("Find(join=%q) err = %v, want %v", tc.join, err, tc.want)
			}
		})
	}
}

func TestJoinValidation_FindTrxReturnsErrorAndSetsTxError(t *testing.T) {
	env := newDosEnv(t)

	err, tx := env.svc.FindTrx(GetAllRequest{Join: "Nope"})
	if !errors.Is(err, ErrInvalidJoin) {
		t.Fatalf("FindTrx err = %v, want ErrInvalidJoin", err)
	}
	if tx.Error == nil {
		t.Fatalf("tx.Error should be set so the query cannot run")
	}
}

func TestJoinValidation_FindOneReturnsSentinel(t *testing.T) {
	env := newDosEnv(t)

	var out dosProduct
	err := env.svc.FindOne(GetAllRequest{Join: "Categories.Products.Categories.Products"}, &out)
	if !errors.Is(err, ErrJoinDepthExceeded) {
		t.Fatalf("FindOne err = %v, want ErrJoinDepthExceeded", err)
	}
}

func TestJoinValidation_DefaultMaxJoinDepthIsThree(t *testing.T) {
	if DefaultMaxJoinDepth != 3 {
		t.Fatalf("DefaultMaxJoinDepth = %d, want 3", DefaultMaxJoinDepth)
	}

	env := newDosEnv(t)

	var out []dosProduct
	var total int64
	if err := env.svc.Find(GetAllRequest{Join: "Categories.Products.Variants"}, &out, &total); err != nil {
		t.Fatalf("3-level join with the default depth failed: %v", err)
	}
	if len(out) != dosProductsCount {
		t.Fatalf("3-level join returned %d rows, want %d", len(out), dosProductsCount)
	}

	err := env.svc.Find(GetAllRequest{Join: "Categories.Products.Categories.Products"}, &out, &total)
	if !errors.Is(err, ErrJoinDepthExceeded) {
		t.Fatalf("4-level join with the default depth err = %v, want ErrJoinDepthExceeded", err)
	}
}

func TestJoinValidation_MaxJoinDepthConfigurable(t *testing.T) {
	t.Run("raised_to_4", func(t *testing.T) {
		env := newDosEnv(t)
		env.svc.Qtb.MaxJoinDepth = 4

		var out []dosProduct
		var total int64
		if err := env.svc.Find(GetAllRequest{Join: "Categories.Products.Variants.Product"}, &out, &total); err != nil {
			t.Fatalf("4-level join with MaxJoinDepth=4 failed: %v", err)
		}
		if len(out) != dosProductsCount {
			t.Fatalf("4-level join returned %d rows, want %d", len(out), dosProductsCount)
		}

		err := env.svc.Find(GetAllRequest{Join: "Categories.Products.Categories.Products.Categories"}, &out, &total)
		if !errors.Is(err, ErrJoinDepthExceeded) {
			t.Fatalf("5-level join with MaxJoinDepth=4 err = %v, want ErrJoinDepthExceeded", err)
		}

		err = env.svc.Find(GetAllRequest{Join: "Categories.Products.Categories.Products"}, &out, &total)
		if !errors.Is(err, ErrInvalidJoin) {
			t.Fatalf("repeated relation with MaxJoinDepth=4 err = %v, want ErrInvalidJoin", err)
		}
	})

	t.Run("lowered_to_2", func(t *testing.T) {
		env := newDosEnv(t)
		env.svc.Qtb.MaxJoinDepth = 2

		var out []dosProduct
		var total int64
		if err := env.svc.Find(GetAllRequest{Join: "Categories.Products"}, &out, &total); err != nil {
			t.Fatalf("2-level join with MaxJoinDepth=2 failed: %v", err)
		}

		err := env.svc.Find(GetAllRequest{Join: "Categories.Products.Categories"}, &out, &total)
		if !errors.Is(err, ErrJoinDepthExceeded) {
			t.Fatalf("3-level join with MaxJoinDepth=2 err = %v, want ErrJoinDepthExceeded", err)
		}
	})
}

func TestJoinValidation_AllowedJoins(t *testing.T) {
	env := newDosEnv(t)
	env.svc.Qtb.AllowedJoins = []string{"categories.products"}

	for _, join := range []string{"Categories", "categories", "Categories.Products"} {
		var out []dosProduct
		var total int64
		if err := env.svc.Find(GetAllRequest{Join: join}, &out, &total); err != nil {
			t.Fatalf("join=%q should be allowed, got %v", join, err)
		}
	}

	var out []dosProduct
	var total int64
	err := env.svc.Find(GetAllRequest{Join: "Variants"}, &out, &total)
	if !errors.Is(err, ErrInvalidJoin) {
		t.Fatalf("join=Variants should be blocked by allowlist, err = %v", err)
	}

	env.svc.Qtb.AllowedJoins = []string{"categories.products.categories"}
	err = env.svc.Find(GetAllRequest{Join: "Categories.Products.Categories"}, &out, &total)
	if !errors.Is(err, ErrInvalidJoin) {
		t.Fatalf("repeated relation must be rejected even when allowlisted, err = %v", err)
	}
}

func TestJoinValidation_WhitespaceAndEmptyPartsAccepted(t *testing.T) {
	env := newDosEnv(t)

	cases := []string{
		" categories . products , variants ",
		",categories",
		"categories,",
	}
	for _, join := range cases {
		var out []dosProduct
		var total int64
		if err := env.svc.Find(GetAllRequest{Join: join}, &out, &total); err != nil {
			t.Fatalf("join=%q should work, got %v", join, err)
		}
		if len(out) != dosProductsCount {
			t.Fatalf("join=%q returned %d rows, want %d", join, len(out), dosProductsCount)
		}
	}
}
