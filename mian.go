package main

import (
	"net/http"
	"github.com/go-ozzo/ozzo-routing"
	"github.com/go-ozzo/ozzo-routing/slash"
	"github.com/HaseProgram/technoparkdb/user"

	"github.com/HaseProgram/technoparkdb/forum"
	"github.com/HaseProgram/technoparkdb/thread"
	"github.com/HaseProgram/technoparkdb/database"
	"github.com/HaseProgram/technoparkdb/post"
	"github.com/HaseProgram/technoparkdb/service"
	_ "net/http/pprof"
)

var router *routing.Router

func main() {
	database.Connect()
	//defer database.DB.Close()
	router := routing.New()
	router.Use(
		slash.Remover(http.StatusMovedPermanently),
	)

	user.Route(router)
	forum.Route(router)
	thread.Route(router)
	post.Route(router)
	service.Route(router)

	http.Handle("/", router)
	http.ListenAndServe(":5000", nil)
}
