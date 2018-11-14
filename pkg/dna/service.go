package dna

import "context"

// Service describes the expected behavior of the DNA sequence service.
// Users can add their DNA, and check if subsequences exist.
type Service interface {
	Add(ctx context.Context, user, token, sequence string) error
	Check(ctx context.Context, user, token, subsequence string) error
}
