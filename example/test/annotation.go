package test

import (
	r "github.com/SphereSoftware/go-annotations/example/test2"
)

type (
	Entity struct {
		Name  string
		Books []Book
	}

	Book struct {
		Name   string
		Price  float32 `default:"1.0"`
		Author *r.Person
	}
)
