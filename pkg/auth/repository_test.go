package auth

import (
	"context"
	"os"
	"testing"
)

func TestSQLiteFixture(t *testing.T) {
	r, err := NewSQLiteRepository("file:testdata/fixture.db")
	if err != nil {
		t.Fatal(err)
	}

	_, err = r.Auth(context.Background(), "bob", "bad password")
	if want, have := ErrBadAuth, err; want != have {
		t.Errorf("Auth with bad creds: want %v, have %v", want, have)
	}
	token, err := r.Auth(context.Background(), "bob", "qwerty")
	if want, have := error(nil), err; want != have {
		t.Fatalf("Auth failed: %v", err)
	}

	if want, have := ErrBadAuth, r.Validate(context.Background(), "bob", "bad token"); want != have {
		t.Errorf("Validate with bad token: want %v, have %v", want, have)
	}
	if want, have := error(nil), r.Validate(context.Background(), "bob", token); want != have {
		t.Errorf("Validate: want %v, have %v", want, have)
	}

	if want, have := ErrBadAuth, r.Deauth(context.Background(), "bob", "bad token"); want != have {
		t.Errorf("Deauth with bad token: want %v, have %v", want, have)
	}
	if want, have := error(nil), r.Deauth(context.Background(), "bob", token); want != have {
		t.Errorf("Deauth: want %v, have %v", want, have)
	}
}

func TestSQLiteIntegration(t *testing.T) {
	var (
		filevar  = "AUTH_INTEGRATION_TEST_FILE"
		filename = os.Getenv(filevar)
	)
	if filename == "" {
		// If a test will write to disk, make it opt-in.
		t.Skipf("skipping; set %s to run this test", filevar)
	}

	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		t.Fatalf("%s: %v", filename, err)
	}

	defer func() {
		if err := os.Remove(filename); err != nil {
			t.Errorf("rm %s: %v", filename, err)
		}
	}()

	r, err := NewSQLiteRepository("file:" + filename)
	if err != nil {
		t.Fatal(err)
	}

	const (
		user = "alpha"
		pass = "beta"
	)
	if want, have := error(nil), r.Create(context.Background(), user, pass); want != have {
		t.Fatalf("Create: want %v, have %v", want, have)
	}

	token, err := r.Auth(context.Background(), user, pass)
	if want, have := error(nil), err; want != have {
		t.Fatalf("Auth: want %v, have %v", want, have)
	}

	if want, have := error(nil), r.Validate(context.Background(), user, token); want != have {
		t.Errorf("Validate: want %v, have %v", want, have)
	}

	if want, have := error(nil), r.Deauth(context.Background(), user, token); want != have {
		t.Errorf("Deauth: want %v, have %v", want, have)
	}
}
