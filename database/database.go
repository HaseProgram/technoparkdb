package database

import (
	"github.com/jackc/pgx"
	"runtime"
	"io/ioutil"
	"fmt"
)

var DB *pgx.ConnPool

func Connect() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	connection := pgx.ConnConfig{
		Host: "localhost",
		User: "hasep",
		Password: "126126",
		Database: "dbproj",
		Port: 5432,
	}

	var err error
	DB, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: connection, MaxConnections: 50})
	if err != nil {
		panic(err)
	}


	rows, qerr := DB.Query("select indexname from pg_indexes")
	if qerr != nil {
		panic(qerr)
	}
	for rows.Next() {
		var idx string
		err = rows.Scan(&idx)
		fmt.Println(idx)
	}
	//err = createShema()
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