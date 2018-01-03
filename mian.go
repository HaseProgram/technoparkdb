package main

import (
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/slash"
	"technoparkdb/user"
	"database/sql"
	_ "github.com/lib/pq"
)

var db *sql.DB

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

	userApi := router.Group("/api/user")
	userApi.Post(`/<nickname:[\w+\.]+>/create`, func(c *routing.Context) error {
		content, responseCode := user.Create(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	userApi.Get(`/<nickname:[\w+\.]+>/profile`, func(c *routing.Context) error {
		content, responseCode := user.Profile(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	userApi.Post(`/<nickname:[\w+\.]+>/profile`, func(c *routing.Context) error {
		content, responseCode := user.Update(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)
}
