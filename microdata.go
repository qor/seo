package seo

import (
	"html/template"
)

type MicroProduct struct {
	Name string
}

type MicroSearch struct {
	URL string
}

type MicroContact struct {
	Telephone string
}

func (MicroProduct) Render() template.HTML {
	return template.HTML(`<span itemprop="name"></span>`)
}

func (MicroSearch) Render() template.HTML {
	return template.HTML("www.example.com")
}

func (MicroContact) Render() template.HTML {
	return template.HTML("86-401-302-313")
}
