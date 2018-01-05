package user

import (
	"github.com/go-ozzo/ozzo-routing"
)

func  Route(router *routing.Router) {
	userApi := router.Group("/api/user")
	userApi.Post(`/<nickname:[\w+\.]+>/create`, func(c *routing.Context) error {
		content, responseCode := Create(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	userApi.Get(`/<nickname:[\w+\.]+>/profile`, func(c *routing.Context) error {
		content, responseCode := Profile(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	userApi.Post(`/<nickname:[\w+\.]+>/profile`, func(c *routing.Context) error {
		content, responseCode := Update(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}

