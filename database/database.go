package database

import (
	"github.com/jackc/pgx"
	"runtime"
)

var DB *pgx.ConnPool

func Connect() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	connection := pgx.ConnConfig{
		Host: "localhost",
		User: "postgres",
		Password: "126126",
		Database: "dbproj",
		Port: 5432,
	}

	var err error
	DB, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: connection, MaxConnections: 10})
	if err != nil {
		panic(err)
	}
}