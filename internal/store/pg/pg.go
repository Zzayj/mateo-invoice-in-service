package pg

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	conn *pgxpool.Pool
}

func NewStore(conn *pgxpool.Pool) *Store {
	return &Store{conn: conn}
}
