package main

import (
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/slash"
	"technoparkdb/user"
	"database/sql"
	_ "github.com/lib/pq"
	"technoparkdb/forum"
)

var db *sql.DB
var router *routing.Router

func main() {
	db, err := sql.Open("postgres", "user=postgres password=126126 dbname=dbproj sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	router := routing.New()
	router.Use(
		slash.Remover(http.StatusMovedPermanently),
	)

	user.Route(router, db)
	forum.Route(router, db)

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)
}
