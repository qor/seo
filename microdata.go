package seo

import (
	"bytes"
	"html/template"
)

// MicroProduct micro product definition, ref: https://developers.google.com/structured-data/rich-snippets/products
type MicroProduct struct {
	Name            string
	Image           string
	Description     string
	BrandName       string
	SKU             string
	RatingValue     float32
	ReviewCount     int
	PriceCurrency   string
	Price           float64
	PriceValidUntil string
	SellerName      string
}

// MicroSearch micro search definition, ref: https://developers.google.com/structured-data/slsb-overview
// e.g.
//   Target: https://query.example-petstore.com/search?q={keyword}
type MicroSearch struct {
	URL        string
	Target     string
	QueryInput string
}

// MicroContact micro search definition, ref: https://developers.google.com/structured-data/customize/contact-points
type MicroContact struct {
	URL         string
	Telephone   string
	ContactType string
}

// Render render micro product structured data
func (product MicroProduct) Render() template.HTML {
	return renderTemplate(productMicroDataTemplate, product)
}

// Render render micro search structured data
func (search MicroSearch) Render() template.HTML {
	return renderTemplate(searchMicroDataTemplate, search)
}

// FormattedQueryInput format query input
func (search MicroSearch) FormattedQueryInput() string {
	if search.QueryInput == "" {
		return "required name=keyword"
	}
	return search.QueryInput
}

// Render render micro contact structured data
func (contact MicroContact) Render() template.HTML {
	return renderTemplate(contactMicroDataTemplate, contact)
}

func renderTemplate(content string, obj interface{}) template.HTML {
	var err error
	if tmpl, err := template.New("").Parse(content); err == nil {
		var results bytes.Buffer
		if err = tmpl.Execute(&results, obj); err == nil {
			return template.HTML(results.String())
		}
		return template.HTML(err.Error())
	}
	return template.HTML(err.Error())
}
