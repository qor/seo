package seo

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
)

var db *gorm.DB
var Admin *admin.Admin
var collection *Collection

func init() {
	db = utils.TestDB()
	db.AutoMigrate(&QorSeoSetting{})
}

// Modal
type SeoGlobalSetting struct {
	SiteName  string
	BrandName string
}

type mircoDataInferface interface {
	Render() template.HTML
}

type Category struct {
	Name string
	SEO  Setting `seo:"type:CategoryPage"`
}

// Test Cases
type RenderTestCase struct {
	SiteName   string
	SeoSetting Setting
	Settings   []interface{}
	Result     string
}

type MicroDataTestCase struct {
	MicroDataType string
	ModelObject   interface{}
	HasTag        string
}

type SeoAppendDefaultValueTestCase struct {
	CatTitle            string
	CatDescription      string
	CatKeywords         string
	CatEnabledCustomize bool
	ExpectTitle         string
	ExpectDescription   string
	ExpectKeywords      string
}

// Runner
func TestRender(t *testing.T) {
	setupSeoCollection()
	category := Category{Name: "Clothing", SEO: Setting{Title: "Using Customize Title", EnabledCustomize: false}}
	categoryWithSeo := Category{Name: "Clothing", SEO: Setting{Title: "Using Customize Title", EnabledCustomize: true}}
	var testCases []RenderTestCase
	testCases = append(testCases,
		// Seo setting are empty
		RenderTestCase{"Qor", Setting{Title: "", Description: "", Keywords: ""}, []interface{}{nil, 123}, `<title></title><meta name="description" content=""><meta name="keywords" content=""/>`},
		// Seo setting have value but variables are emptry
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}}", Description: "{{SiteName}}", Keywords: "{{SiteName}}"}, []interface{}{"", ""}, `<title>Qor</title><meta name="description" content="Qor"><meta name="keywords" content="Qor"/>`},
		// Seo setting change Site Name
		RenderTestCase{"ThePlant Qor", Setting{Title: "{{SiteName}}", Description: "{{SiteName}}", Keywords: "{{SiteName}}"}, []interface{}{"", ""}, `<title>ThePlant Qor</title><meta name="description" content="ThePlant Qor"><meta name="keywords" content="ThePlant Qor"/>`},
		// Seo setting have value and variables are present
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}} {{Name}}", Description: "{{URLTitle}}", Keywords: "{{URLTitle}}"}, []interface{}{"Clothing", "/clothing"}, `<title>Qor Clothing</title><meta name="description" content="/clothing"><meta name="keywords" content="/clothing"/>`},
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}} {{Name}} {{Name}}", Description: "{{URLTitle}} {{URLTitle}}", Keywords: "{{URLTitle}} {{URLTitle}}"}, []interface{}{"Clothing", "/clothing"}, `<title>Qor Clothing Clothing</title><meta name="description" content="/clothing /clothing"><meta name="keywords" content="/clothing /clothing"/>`},
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}} {{Name}} {{URLTitle}}", Description: "{{SiteName}} {{Name}} {{URLTitle}}", Keywords: "{{SiteName}} {{Name}} {{URLTitle}}"}, []interface{}{"", ""}, `<title>Qor  </title><meta name="description" content="Qor  "><meta name="keywords" content="Qor  "/>`},
		// Using undefined variables
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}} {{Name1}}", Description: "{{URLTitle1}}", Keywords: "{{URLTitle1}}"}, []interface{}{"Clothing", "/clothing"}, `<title>Qor </title><meta name="description" content=""><meta name="keywords" content=""/>`},
		// Using Resource's seo
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}}", Description: "{{URLTitle}}", Keywords: "{{URLTitle}}"}, []interface{}{category}, `<title>Qor</title><meta name="description" content=""><meta name="keywords" content=""/>`},
		RenderTestCase{"Qor", Setting{Title: "{{SiteName}}", Description: "{{URLTitle}}", Keywords: "{{URLTitle}}"}, []interface{}{categoryWithSeo}, `<title>Using Customize Title</title><meta name="description" content=""><meta name="keywords" content=""/>`},
	)
	i := 1
	context := &qor.Context{DB: db}
	for _, testCase := range testCases {
		createGlobalSetting(testCase.SiteName)
		createCategoryPageSetting(testCase.SeoSetting)
		metatHTML := string(collection.Render(context, "CategoryPage", testCase.Settings...))
		metatHTML = strings.Replace(metatHTML, "\n", "", -1)
		if string(metatHTML) == testCase.Result {
			color.Green(fmt.Sprintf("Seo Render TestCase #%d: Success\n", i))
		} else {
			t.Errorf(color.RedString(fmt.Sprintf("\nSeo Render TestCase #%d: Failure Result:%s\n", i, string(metatHTML))))
		}
		i++
	}
}

func TestSeoSections(t *testing.T) {
	setupSeoCollection()
	var count int
	db.Model(QorSeoSetting{}).Count(&count)
	if count != 0 {
		t.Errorf(color.RedString("\nSeoSections TestCase #1: should get empty settings"))
	}

	seoSections(&admin.Context{Context: &qor.Context{DB: db}}, collection)
	db.Model(QorSeoSetting{}).Count(&count)
	if count != 2 {
		t.Errorf(color.RedString("\nSeoSections TestCase #2: should get two settings"))
	}

	var settings []QorSeoSetting
	settingNames := []string{"CategoryPage", "DefaultPage"}
	db.Model(QorSeoSetting{}).Order("Name ASC").Find(&settings)
	for i, setting := range settings {
		if setting.Name != settingNames[i] {
			t.Errorf(color.RedString(fmt.Sprintf("\nSeoSections TestCase #%v: should has setting `%v`", 3+i, settingNames[i])))
		}
	}
}

func TestSeoGlobalSetting(t *testing.T) {
	setupSeoCollection()
	var count int
	db.Model(QorSeoSetting{}).Count(&count)
	if count != 0 {
		t.Errorf(color.RedString("\nSeoGlobalSetting TestCase #1: global setting should be empty"))
	}

	seoGlobalSetting(&admin.Context{Context: &qor.Context{DB: db}}, collection)
	var settings []QorSeoSetting
	db.Find(&settings)
	if len(settings) != 1 || !settings[0].IsGlobalSeo {
		t.Errorf(color.RedString("\nSeoGlobalSetting TestCase #2: global setting should be present"))
	}
}

func TestSeoGlobalSettingValue(t *testing.T) {
	setupSeoCollection()
	setting := createGlobalSetting("New Site Name")
	globalSetting := seoGlobalSettingValue(collection, setting)
	if globalSetting.(SeoGlobalSetting).SiteName != "New Site Name" {
		t.Errorf(color.RedString("\nSeoGlobalSettingValue TestCase #1: value doesn't be set"))
	}
}

func TestSeoAppendDefaultValue(t *testing.T) {
	setupSeoCollection()
	createCategoryPageSetting(Setting{Title: "GT", Description: "GD", Keywords: "GK"})
	testCases := []SeoAppendDefaultValueTestCase{
		{CatTitle: "T", CatDescription: "D", CatKeywords: "K", CatEnabledCustomize: true, ExpectTitle: "T", ExpectDescription: "D", ExpectKeywords: "K"},
		{CatTitle: "T", CatDescription: "D", CatKeywords: "K", CatEnabledCustomize: false, ExpectTitle: "T", ExpectDescription: "D", ExpectKeywords: "K"},
		{CatTitle: "", CatDescription: "", CatKeywords: "", CatEnabledCustomize: true, ExpectTitle: "", ExpectDescription: "", ExpectKeywords: ""},
		{CatTitle: "", CatDescription: "", CatKeywords: "", CatEnabledCustomize: false, ExpectTitle: "GT", ExpectDescription: "GD", ExpectKeywords: "GK"},
	}
	for i, testCase := range testCases {
		category := Category{SEO: Setting{Title: testCase.CatTitle, Description: testCase.CatDescription, Keywords: testCase.CatKeywords, EnabledCustomize: testCase.CatEnabledCustomize}}
		seo := collection.GetSeo("CategoryPage")
		setting := seoAppendDefaultValue(&admin.Context{Context: &qor.Context{DB: db}}, seo, category.SEO).(Setting)
		var hasError bool
		if setting.Title != testCase.ExpectTitle {
			hasError = true
			t.Errorf(color.RedString("\nSeoAppendDefaultValue TestCase #%v: title should be equal %v, but got %v", i+1, testCase.ExpectTitle, setting.Title))
		}
		if setting.Description != testCase.ExpectDescription {
			hasError = true
			t.Errorf(color.RedString("\nSeoAppendDefaultValue TestCase #%v: description should be equal %v, but got %v", i+1, testCase.ExpectDescription, setting.Description))
		}
		if setting.Keywords != testCase.ExpectKeywords {
			hasError = true
			t.Errorf(color.RedString("\nSeoAppendDefaultValue TestCase #%v: keywords should be equal %v, but go %v", i+1, testCase.ExpectKeywords, setting.Keywords))
		}
		if !hasError {
			color.Green(fmt.Sprintf("SeoAppendDefaultValue TestCase #%v: Success", i+1))
		}
	}
}

func TestSeoTagsByType(t *testing.T) {
	setupSeoCollection()
	seo := collection.GetSeo("CategoryPage")
	validTags := seoTagsByType(seo)
	tags := []string{"SiteName", "BrandName", "Name", "URLTitle"}
	if strings.Join(validTags, ",") != strings.Join(tags, ",") {
		t.Errorf(color.RedString("\nSeoTagsByType TestCase: seo's tags should be %v", tags))
	}
}

func TestMicrodata(t *testing.T) {
	var testCases []MicroDataTestCase
	testCases = append(testCases,
		MicroDataTestCase{"Product", MicroProduct{Name: ""}, `<span itemprop="name"></span>`},
		MicroDataTestCase{"Product", MicroProduct{Name: "Polo"}, `<span itemprop="name">Polo</span>`},
		MicroDataTestCase{"Search", MicroSearch{Target: "http://www.example.com/q={keyword}"}, `http:\/\/www.example.com\/q={keyword}`},
		MicroDataTestCase{"Contact", MicroContact{Telephone: "86-401-302-313"}, `86-401-302-313`},
	)
	i := 1
	for _, microDataTestCase := range testCases {
		tagHTML := reflect.ValueOf(microDataTestCase.ModelObject).Interface().(mircoDataInferface).Render()
		if strings.Contains(string(tagHTML), microDataTestCase.HasTag) {
			color.Green(fmt.Sprintf("Seo Micro TestCase #%d: Success\n", i))
		} else {
			t.Errorf(color.RedString(fmt.Sprintf("\nSeo Micro TestCase #%d: Failure Result:%s\n", i, string(tagHTML))))
		}
		i++
	}
}

// Created related methods
func setupSeoCollection() {
	if err := db.DropTableIfExists(&QorSeoSetting{}).Error; err != nil {
		panic(err)
	}
	db.AutoMigrate(&QorSeoSetting{})
	collection = New("Seo")
	collection.RegisterGlobalVaribles(&SeoGlobalSetting{SiteName: "Qor SEO", BrandName: "Qor"})
	collection.RegisterSeo(&SEO{
		Name: "DefaultPage",
	})
	collection.RegisterSeo(&SEO{
		Name:     "CategoryPage",
		Varibles: []string{"Name", "URLTitle"},
		Context: func(objects ...interface{}) (context map[string]string) {
			context = make(map[string]string)
			if len(objects) > 0 && objects[0] != nil {
				if v, ok := objects[0].(string); ok {
					context["Name"] = v
				}
			}
			if len(objects) > 1 && objects[1] != nil {
				if v, ok := objects[1].(string); ok {
					context["URLTitle"] = v
				}
			}
			return context
		},
	})
	Admin = admin.New(&qor.Config{DB: db})
	Admin.AddResource(collection, &admin.Config{Name: "SEO Setting", Menu: []string{"Site Management"}, Singleton: true})
	Admin.MountTo("/admin", http.NewServeMux())
}

func createGlobalSetting(siteName string) *QorSeoSetting {
	globalSeoSetting := QorSeoSetting{}
	db.Where("name = ?", "Seo").Find(&globalSeoSetting)
	globalSetting := make(map[string]string)
	globalSetting["SiteName"] = siteName
	globalSeoSetting.Setting = Setting{GlobalSetting: globalSetting}
	globalSeoSetting.Name = "Seo"
	globalSeoSetting.IsGlobalSeo = true
	if db.NewRecord(globalSeoSetting) {
		db.Create(&globalSeoSetting)
	} else {
		db.Save(&globalSeoSetting)
	}
	return &globalSeoSetting
}

func createCategoryPageSetting(setting Setting) {
	seoSetting := QorSeoSetting{}
	db.Where("name = ?", "CategoryPage").First(&seoSetting)
	seoSetting.Setting = setting
	seoSetting.Name = "CategoryPage"
	if db.NewRecord(seoSetting) {
		db.Create(&seoSetting)
	} else {
		db.Save(&seoSetting)
	}
}
