package repository

import (
	"github.com/BloggingApp/notification-service/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Postgres *postgres.PGRepo
}

func New(db *pgxpool.Conn) *Repository {
	return &Repository{
		Postgres: postgres.New(db),
	}
}
