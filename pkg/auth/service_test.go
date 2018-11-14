package auth

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestFlow(t *testing.T) {
	s := NewDefaultService(newMockRepo())

	if want, have := error(nil), s.Signup(context.Background(), "peter", "123456"); want != have {
		t.Fatalf("Signup: want %v, have %v", want, have)
	}

	token, err := s.Login(context.Background(), "peter", "123456")
	if want, have := error(nil), err; want != have {
		t.Fatalf("Login: want %v, have %v", want, have)
	}

	if want, have := error(nil), s.Validate(context.Background(), "peter", token); want != have {
		t.Errorf("Validate: want %v, have %v", want, have)
	}

	if want, have := error(nil), s.Logout(context.Background(), "peter", token); want != have {
		t.Errorf("Logout: want %v, have %v", want, have)
	}

	if want, have := ErrBadAuth, s.Validate(context.Background(), "peter", token); want != have {
		t.Errorf("Validate after Logout: want %v, have %v", want, have)
	}
}

type mockRepo struct {
	creds  map[string]string
	tokens map[string]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		creds:  map[string]string{},
		tokens: map[string]string{},
	}
}

func (r *mockRepo) Create(ctx context.Context, user, pass string) error {
	if _, ok := r.creds[user]; ok {
		return errors.New("user already exists")
	}

	r.creds[user] = pass
	return nil
}

func (r *mockRepo) Auth(ctx context.Context, user, pass string) (token string, err error) {
	if have, ok := r.creds[user]; !ok || pass != have {
		return "", ErrBadAuth
	}

	p := make([]byte, 8)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(p)
	token = fmt.Sprintf("%x", p)
	r.tokens[user] = token
	return token, nil
}

func (r *mockRepo) Deauth(ctx context.Context, user, token string) error {
	if have, ok := r.tokens[user]; !ok || token != have {
		return ErrBadAuth
	}
	delete(r.tokens, user)
	return nil
}

func (r *mockRepo) Validate(ctx context.Context, user, token string) error {
	if have, ok := r.tokens[user]; !ok || token != have {
		return ErrBadAuth
	}

	return nil
}
