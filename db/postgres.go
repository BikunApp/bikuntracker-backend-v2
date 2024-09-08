package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool(dsn string) (pool *pgxpool.Pool) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to create database pool: %s", err.Error())
	}
	return pool
}

func TestConnection(pool *pgxpool.Pool) {
	ctx := context.Background()
	_, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err.Error())
	}
	fmt.Println("connected to database successfully")
}
