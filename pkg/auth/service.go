package auth

import (
	"context"
	"errors"
)

// Service describes the expected behavior of the authentication service.
// Users can create accounts, log in, and log out; other services can
// validate user tokens (sessions) that they've received.
type Service interface {
	Signup(ctx context.Context, user, pass string) error
	Login(ctx context.Context, user, pass string) (token string, err error)
	Logout(ctx context.Context, user, token string) error
	Validate(ctx context.Context, user, token string) error
}

// ErrBadAuth is returned when authentication fails for any reason.
var ErrBadAuth = errors.New("bad auth")
