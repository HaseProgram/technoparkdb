package forum

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
)

func  Route(router *routing.Router, db *sql.DB) {
	forumApi := router.Group("/api/forum")
	forumApi.Post(`/create`, func(c *routing.Context) error {
		content, responseCode := Create(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	forumApi.Get(`/<slug:[\w+\.\-\_]+>/details`, func(c *routing.Context) error {
		content, responseCode := Details(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}