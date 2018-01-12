package thread

import (
	"github.com/go-ozzo/ozzo-routing"
	"github.com/HaseProgram/technoparkdb/post"
)

func  Route(router *routing.Router) {
	threadApi := router.Group("/api/thread")
	threadApi.Get(`/<slugid:[\w+\.\-\_]+>/details`, func(c *routing.Context) error {
		content, responseCode := Details(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	threadApi.Post(`/<slugid:[\w+\.\-\_]+>/details`, func(c *routing.Context) error {
		content, responseCode := Update(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	threadApi.Post(`/<slugid:[\w+\.\-\_]+>/create`, func(c *routing.Context) error {
		content, responseCode := post.Create(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	threadApi.Get(`/<slugid:[\w+\.\-\_]+>/posts`, func(c *routing.Context) error {
		content, responseCode := post.GetPosts(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})

	threadApi.Post(`/<slugid:[\w+\.\-\_]+>/vote`, func(c *routing.Context) error {
		content, responseCode := Vote(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}