package main

import (
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/slash"
	"technoparkdb/user"

	"technoparkdb/forum"
	"technoparkdb/thread"
	"technoparkdb/database"
	"technoparkdb/post"
)

var router *routing.Router

func main() {
	database.Connect()
	defer database.DB.Close()
	router := routing.New()
	router.Use(
		slash.Remover(http.StatusMovedPermanently),
	)

	user.Route(router)
	forum.Route(router)
	thread.Route(router)
	post.Route(router)

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)
}
