package thread

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
)

func  Route(router *routing.Router, db *sql.DB) {
	threadApi := router.Group("/api/thread")
	threadApi.Get(`/<slugid:[\w+\.\-\_]+>/details`, func(c *routing.Context) error {
		content, responseCode := Details(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	threadApi.Post(`/<slugid:[\w+\.\-\_]+>/details`, func(c *routing.Context) error {
		content, responseCode := Update(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}