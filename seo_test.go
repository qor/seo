package seo_test

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/test/utils"
	"github.com/qor/seo"
)

var db *gorm.DB
var seoCollection *seo.SeoCollection

func init() {
	db = utils.TestDB()
	db.AutoMigrate(&seo.QorSeoSetting{})
}

// Modal
type SeoGlobalSetting struct {
	SiteName  string
	BrandName string
}

type mircoDataInferface interface {
	Render() template.HTML
}

// Test Cases
type RenderTestCase struct {
	SiteName         string
	SeoSetting       seo.Setting
	CategoryName     interface{}
	CategoryURLTitle interface{}
	Result           string
}

type MicroDataTestCase struct {
	MicroDataType string
	ModelObject   interface{}
	HasTag        string
}

// Runner
func TestRender(t *testing.T) {
	setupSeoCollection()
	var testCases []RenderTestCase
	testCases = append(testCases,
		// Seo setting are empty
		RenderTestCase{"Qor", seo.Setting{Title: "", Description: "", Keywords: ""}, nil, nil, `<title></title><meta name="description" content=""><meta name="keywords" content=""/>`},
		// Seo setting have value but variables are emptry
		RenderTestCase{"Qor", seo.Setting{Title: "{{SiteName}}", Description: "{{SiteName}}", Keywords: "{{SiteName}}"}, "", "", `<title>Qor</title><meta name="description" content="Qor"><meta name="keywords" content="Qor"/>`},
		// Seo setting have value and variables are present
		RenderTestCase{"Qor", seo.Setting{Title: "{{SiteName}} {{Name}}", Description: "{{URLTitle}}", Keywords: "{{URLTitle}}"}, "Clothing", "/clothing", `<title>Qor Clothing</title><meta name="description" content="/clothing"><meta name="keywords" content="/clothing"/>`},
		RenderTestCase{"Qor", seo.Setting{Title: "{{SiteName}} {{Name}} {{Name}}", Description: "{{URLTitle}} {{URLTitle}}", Keywords: "{{URLTitle}} {{URLTitle}}"}, "Clothing", "/clothing", `<title>Qor Clothing Clothing</title><meta name="description" content="/clothing /clothing"><meta name="keywords" content="/clothing /clothing"/>`},
		RenderTestCase{"Qor", seo.Setting{Title: "{{SiteName}} {{Name}} {{URLTitle}}", Description: "{{SiteName}} {{Name}} {{URLTitle}}", Keywords: "{{SiteName}} {{Name}} {{URLTitle}}"}, "", "", `<title>Qor  </title><meta name="description" content="Qor  "><meta name="keywords" content="Qor  "/>`},
		// Using undefined variables
		RenderTestCase{"Qor", seo.Setting{Title: "{{SiteName}} {{Name1}}", Description: "{{URLTitle1}}", Keywords: "{{URLTitle1}}"}, "Clothing", "/clothing", `<title>Qor </title><meta name="description" content=""><meta name="keywords" content=""/>`},
	)
	i := 1
	for _, testCase := range testCases {
		createSetting(testCase.SiteName, testCase.SeoSetting)
		metatHTML := string(seoCollection.Render("CategoryPage", testCase.CategoryName, testCase.CategoryURLTitle))
		metatHTML = strings.Replace(metatHTML, "\n", "", -1)
		if string(metatHTML) == testCase.Result {
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

func setupSeoCollection() {
	seoCollection = seo.New()
	seoCollection.RegisterGlobalSetting(&SeoGlobalSetting{SiteName: "Qor SEO", BrandName: "Qor"})
	seoCollection.RegisterSeo(&seo.Seo{
		Name: "Default Page",
	})
	seoCollection.RegisterSeo(&seo.Seo{
		Name:     "CategoryPage",
		Settings: []string{"Name", "URLTitle"},
		Context: func(objects ...interface{}) (context map[string]string) {
			context = make(map[string]string)
			if len(objects) > 0 && objects[0] != nil {
				context["Name"] = objects[0].(string)
			}
			if len(objects) > 1 && objects[1] != nil {
				context["URLTitle"] = objects[1].(string)
			}
			return context
		},
	})
	Admin := admin.New(&qor.Config{DB: db})
	Admin.AddResource(seoCollection, &admin.Config{Name: "SEO Setting", Menu: []string{"Site Management"}, Singleton: true})
	Admin.MountTo("/admin", http.NewServeMux())
}

func createSetting(siteName string, setting seo.Setting) {
	globalSeoSetting := seo.QorSeoSetting{}
	globalSetting := make(map[string]string)
	globalSetting["SiteName"] = siteName
	globalSeoSetting.Setting = seo.Setting{GlobalSetting: globalSetting}
	globalSeoSetting.Name = "QorSeoGlobalSettings"
	db.Where("name = ?", "QorSeoGlobalSettings").Find(&globalSeoSetting)
	if db.NewRecord(globalSeoSetting) {
		db.Create(&globalSeoSetting)
	} else {
		db.Update(&globalSeoSetting)
	}

	seoSetting := seo.QorSeoSetting{}
	db.Where("name = ?", "CategoryPage").First(&seoSetting)
	seoSetting.Setting = setting
	if db.NewRecord(seoSetting) {
		db.Create(&seoSetting)
	} else {
		db.Save(&seoSetting)
	}
}
