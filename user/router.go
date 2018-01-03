package user

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
)

func  Route(router *routing.Router, db *sql.DB) {
	userApi := router.Group("/api/user")
	userApi.Post(`/<nickname:[\w+\.]+>/create`, func(c *routing.Context) error {
		content, responseCode := Create(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	userApi.Get(`/<nickname:[\w+\.]+>/profile`, func(c *routing.Context) error {
		content, responseCode := Profile(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	userApi.Post(`/<nickname:[\w+\.]+>/profile`, func(c *routing.Context) error {
		content, responseCode := Update(c, db)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}

