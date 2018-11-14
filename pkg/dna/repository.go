package dna

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // driver
	"github.com/pkg/errors"
)

// ErrInvalidUser is returned when an invalid user is passed to Select.
var ErrInvalidUser = errors.New("invalid user")

// SQLiteRepository for persistence of the DNA sequences.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository connects to the DB represented by URN.
func NewSQLiteRepository(urn string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", urn)
	if err != nil {
		return nil, errors.Wrap(err, "error opening DB")
	}

	if _, err := db.Query(`SELECT 1 FROM dna`); err != nil {
		if _, err := db.Exec(`CREATE TABLE dna (user STRING NOT NULL PRIMARY KEY, sequence STRING NOT NULL)`); err != nil {
			return nil, errors.Wrap(err, "error creating dna table")
		}
		for user, sequence := range map[string]string{
			"alice": "attcgtattattttttgatatttttccacaaaaatacagactaaatacaactgaatacag",
			"bob":   "tgcaaaattagatataaatgtaaacgaacataaaaacttttataagacaggattaagtta",
		} {
			if _, err := db.Exec(`INSERT INTO dna (user, sequence) VALUES (?, ?)`, user, sequence); err != nil {
				return nil, errors.Wrap(err, "error populating initial sequences")
			}
		}
	}

	return &SQLiteRepository{
		db: db,
	}, nil
}

// Insert a user's DNA sequence to the repository.
func (r *SQLiteRepository) Insert(ctx context.Context, user, sequence string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO dna(user, sequence) VALUES(?, ?)`, user, sequence)
	if err != nil {
		return errors.Wrap(err, "error writing to repository")
	}
	return nil
}

// Select a user's DNA sequence from the repository.
func (r *SQLiteRepository) Select(ctx context.Context, user string) (sequence string, err error) {
	if err := r.db.QueryRowContext(ctx, `SELECT sequence FROM dna WHERE user = ?`, user).Scan(&sequence); err == sql.ErrNoRows {
		return "", ErrInvalidUser
	} else if err != nil {
		return "", errors.Wrap(err, "error reading from repository")
	}
	return sequence, nil
}
