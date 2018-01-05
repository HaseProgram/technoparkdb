package post

import (
	"github.com/go-ozzo/ozzo-routing"
)

func  Route(router *routing.Router) {
	threadApi := router.Group("/api/post")
	threadApi.Get(`/<id:\d+>/details`, func(c *routing.Context) error {
		content, responseCode := Details(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	threadApi.Post(`/<id:\d+>/details`, func(c *routing.Context) error {
		content, responseCode := Update(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}