package seo_test

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/test/utils"
	"github.com/qor/seo"
	"strings"
	"testing"
)

var db *gorm.DB

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

func init() {
	db = utils.TestDB()
	db.AutoMigrate(&Seo{}, &Category{})
}

type RenderTestCase struct {
	SiteName         string
	HomePage         seo.Setting
	CategoryName     string
	CategoryURLTitle string
	Result           string
}

func TestRender(t *testing.T) {
	var testCases []RenderTestCase
	testCases = append(testCases,
		RenderTestCase{"Qor", seo.Setting{"", "", "Name,URLTitle", nil}, "", "", `<title></title><meta name="description" content="">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}}", "{{SiteName}}", "Name,URLTitle", nil}, "", "", `<title>Qor</title><meta name="description" content="Qor">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}}", "{{URLTitle}}", "Name,URLTitle", nil}, "Clothing", "/clothing", `<title>Qor Clothing</title><meta name="description" content="/clothing">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}} {{Name}}", "{{URLTitle}} {{URLTitle}}", "Name,URLTitle", nil}, "Clothing", "/clothing", `<title>Qor Clothing Clothing</title><meta name="description" content="/clothing /clothing">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name}} {{URLTitle}}", "{{SiteName}} {{Name}} {{URLTitle}}", "Name,URLTitle", nil}, "", "", `<title>Qor  </title><meta name="description" content="Qor  ">`},
		RenderTestCase{"Qor", seo.Setting{"{{SiteName}} {{Name1}}", "{{URLTitle1}}", "Name,URLTitle", nil}, "Clothing", "/clothing", `<title>Qor {{Name1}}</title><meta name="description" content="{{URLTitle1}}">`},
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
		metatHTML := string(seo.HomePage.Render(seo, cat))
		metatHTML = strings.Replace(metatHTML, "\n", "", -1)
		if string(metatHTML) == renderTestCase.Result {
			color.Green(fmt.Sprintf("Seo Render TestCase #%d: Success\n", i))
		} else {
			t.Errorf(color.RedString(fmt.Sprintf("\nSeo Render TestCase #%d: Failure Result:%s\n", i, string(metatHTML))))
		}
		i += 1
	}
}
