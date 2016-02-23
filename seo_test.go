package seo_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/test/utils"
	"github.com/qor/seo"
	"html/template"
	"reflect"
	"strings"
	"testing"
)

var db *gorm.DB

func init() {
	db = utils.TestDB()
	db.AutoMigrate(&Seo{}, &Category{})
}

// Modal
type Seo struct {
	gorm.Model
	SiteName string
	HomePage seo.Setting `seo:"Name,URLTitle"`
}

type Category struct {
	gorm.Model
	Name     string
	URLTitle string
}

type mircoDataInferface interface {
	Render() template.HTML
}

// Test Cases
type RenderTestCase struct {
	SiteName         string
	HomePage         seo.Setting
	CategoryName     string
	CategoryURLTitle string
	IsPassNonStruct  bool
	Result           string
}

type MicroDataTestCase struct {
	MicroDataType string
	ModelObject   interface{}
	HasTag        string
}

// Runner
func TestRender(t *testing.T) {
	var testCases []RenderTestCase
	testCases = append(testCases,
		RenderTestCase{"Qor", seo.Setting{"", "", "Name,URLTitle", nil}, "", "", false, `<title></title><meta name="description" content="">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}}", "{{SiteName}}", "Name,URLTitle", nil}, "", "", false, `<title>Qor</title><meta name="description" content="Qor">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}}", "{{URLTitle}}", "Name,URLTitle", nil}, "Clothing", "/clothing", false, `<title>Qor Clothing</title><meta name="description" content="/clothing">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}} {{Name}}", "{{URLTitle}} {{URLTitle}}", "Name,URLTitle", nil}, "Clothing", "/clothing", false, `<title>Qor Clothing Clothing</title><meta name="description" content="/clothing /clothing">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}} {{URLTitle}}", "{{SiteName}} {{Name}} {{URLTitle}}", "Name,URLTitle", nil}, "", "", false, `<title>Qor  </title><meta name="description" content="Qor  ">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name1}}", "{{URLTitle1}}", "Name,URLTitle", nil}, "Clothing", "/clothing", false, `<title>Qor {{Name1}}</title><meta name="description" content="{{URLTitle1}}">`},
		// Pass nil object for Render
		RenderTestCase{"Qor", seo.Setting{}, "Clothing", "/clothing", false, `<title></title><meta name="description" content="">`},
		// Pass a non struct object
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}}", "{{URLTitle}}", "Name,URLTitle", nil}, "Clothing", "/clothing", true, `<title>{{SiteName}} {{Name}}</title><meta name="description" content="{{URLTitle}}">`},
	)
	seo := Seo{}
	cat := Category{}
	i := 1
	for _, renderTestCase := range testCases {
		seo.SiteName = renderTestCase.SiteName
		seo.HomePage = renderTestCase.HomePage
		db.Save(&seo)
		cat.Name = renderTestCase.CategoryName
		cat.URLTitle = renderTestCase.CategoryURLTitle
		db.Save(&cat)
		var metatHTML string
		if seo.HomePage.Title == "" && seo.HomePage.Description == "" {
			metatHTML = string(seo.HomePage.Render(seo, nil))
		} else {
			metatHTML = string(seo.HomePage.Render(seo, cat))
		}
		if renderTestCase.IsPassNonStruct {
			metatHTML = string(seo.HomePage.Render(true, true))
		}
		metatHTML = strings.Replace(metatHTML, "\n", "", -1)
		if string(metatHTML) == renderTestCase.Result {
			color.Green(fmt.Sprintf("Seo Render TestCase #%d: Success\n", i))
		} else {
			t.Errorf(color.RedString(fmt.Sprintf("\nSeo Render TestCase #%d: Failure Result:%s\n", i, string(metatHTML))))
		}
		i += 1
	}
}

func TestMicrodata(t *testing.T) {
	var testCases []MicroDataTestCase
	testCases = append(testCases,
		MicroDataTestCase{"Product", seo.MicroProduct{Name: ""}, `<span itemprop="name"></span>`},
		MicroDataTestCase{"Product", seo.MicroProduct{Name: "Polo"}, `<span itemprop="name">Polo</span>`},
		MicroDataTestCase{"Search", seo.MicroSearch{Target: "http://www.example.com/q={keyword}"}, `http:\/\/www.example.com\/q={keyword}`},
		MicroDataTestCase{"Contact", seo.MicroContact{Telephone: "86-401-302-313"}, `86-401-302-313`},
	)
	i := 1
	for _, microDataTestCase := range testCases {
		tagHTML := reflect.ValueOf(microDataTestCase.ModelObject).Interface().(mircoDataInferface).Render()
		if strings.Contains(string(tagHTML), microDataTestCase.HasTag) {
			color.Green(fmt.Sprintf("Seo Micro TestCase #%d: Success\n", i))
		} else {
			t.Errorf(color.RedString(fmt.Sprintf("\nSeo Micro TestCase #%d: Failure Result:%s\n", i, string(tagHTML))))
		}
		i += 1
	}
}
