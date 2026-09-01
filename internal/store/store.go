package store

import "github.com/jackc/pgx/v5/pgxpool"

// Store wraps the connection pool and exposes data-access methods for each model.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }
