package service

import (
"github.com/go-ozzo/ozzo-routing"
)

func  Route(router *routing.Router) {
	serviceApi := router.Group("/api/service")
	serviceApi.Get(`/status`, func(c *routing.Context) error {
		content, responseCode := Status(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
	serviceApi.Post(`/clear`, func(c *routing.Context) error {
		content, responseCode := Clear(c)
		c.Response.Header().Set("Content-Type", "application/json")
		c.Response.WriteHeader(responseCode)
		return c.Write(content)
	})
}