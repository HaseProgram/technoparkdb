package database

import (
	"github.com/jackc/pgx"
	"runtime"
	"io/ioutil"
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
	DB, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: connection, MaxConnections: 50})
	if err != nil {
		panic(err)
	}
	err = createShema()
	if err != nil {
		panic(err)
	}
}

func createShema() error {
	sql, err := ioutil.ReadFile("db.sql")
	if err != nil {
		return err
	}
	shema := string(sql)

	_, err = DB.Exec(shema)
	return err
}