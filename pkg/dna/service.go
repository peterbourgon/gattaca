package dna

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

// Service describes the expected behavior of the DNA sequence service.
// Users can add their DNA, and check if subsequences exist.
type Service interface {
	Add(ctx context.Context, user, token, sequence string) error
	Check(ctx context.Context, user, token, subsequence string) error
}

// DefaultService provides our DNA sequence business logic.
type DefaultService struct {
	repo  Repository
	valid Validator
}

var (
	// ErrSubsequenceNotFound is returned by Check on a failure.
	ErrSubsequenceNotFound = errors.New("subsequence doesn't appear in the DNA sequence")

	// ErrBadAuth is returned if a user validation check fails.
	ErrBadAuth = errors.New("bad auth")

	// ErrInvalidSequence is returned if an invalid sequence is added.
	ErrInvalidSequence = errors.New("invalid DNA sequence")
)

// Repository is a client-side interface, which models
// the concrete e.g. SQLiteRepository.
type Repository interface {
	Insert(ctx context.Context, user, sequence string) error
	Select(ctx context.Context, user string) (sequence string, err error)
}

// Validator is a client-side interface, which models
// the parts of the auth service that we use.
type Validator interface {
	Validate(ctx context.Context, user, token string) error
}

// NewDefaultService returns a usable service, wrapping a repository.
func NewDefaultService(r Repository, v Validator) *DefaultService {
	return &DefaultService{
		repo:  r,
		valid: v,
	}
}

// Add a user and their DNA sequence to the database.
func (s *DefaultService) Add(ctx context.Context, user, token, sequence string) (err error) {
	if err := s.valid.Validate(ctx, user, token); err != nil {
		return ErrBadAuth
	}

	if !validSequence(sequence) {
		return ErrInvalidSequence
	}

	if err := s.repo.Insert(ctx, user, sequence); err != nil {
		return errors.Wrap(err, "error adding new user")
	}

	return nil
}

// Check returns true if the given subsequence is present in the user's DNA.
func (s *DefaultService) Check(ctx context.Context, user, token, subsequence string) (err error) {
	if err := s.valid.Validate(ctx, user, token); err != nil {
		return ErrBadAuth
	}

	sequence, err := s.repo.Select(ctx, user)
	if err != nil {
		return errors.Wrap(err, "error reading DNA sequence from repository")
	}

	if !strings.Contains(sequence, subsequence) {
		return ErrSubsequenceNotFound
	}

	return nil
}

func validSequence(sequence string) bool {
	for _, r := range sequence {
		switch r {
		case 'g', 'a', 't', 'c':
			continue
		default:
			return false
		}
	}
	return true
}
