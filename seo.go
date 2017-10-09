package seo

import (
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

func init() {
	admin.RegisterViewPath("github.com/qor/seo/views")
}

// New initialize a SeoCollection instance
func New(name string) *Collection {
	return &Collection{Name: name}
}

// Collection will hold registered seo configures and global setting definition and other configures
type Collection struct {
	Name            string
	SettingResource *admin.Resource

	registeredSEO  []*SEO
	resource       *admin.Resource
	globalResource *admin.Resource
	globalSetting  interface{}
}

// SEO represents a seo object for a page
type SEO struct {
	Name       string
	Varibles   []string
	Context    func(...interface{}) map[string]string
	collection *Collection
}

// RegisterGlobalVaribles register global setting struct and will represents as 'Site-wide Settings' part in admin
func (collection *Collection) RegisterGlobalVaribles(s interface{}) {
	collection.globalSetting = s
}

// RegisterSEO register a seo
func (collection *Collection) RegisterSEO(seo *SEO) {
	seo.collection = collection
	collection.registeredSEO = append(collection.registeredSEO, seo)
}

// Render render SEO Setting
func (collection Collection) Render(context *qor.Context, name string, objects ...interface{}) template.HTML {
	var (
		title, description, keywords string
		seoField                     Setting
		db                           = context.GetDB()
		seo                          = collection.GetSEO(name)
	)

	// If passed objects has customzied SEO Setting field
	for _, obj := range objects {
		if value := reflect.Indirect(reflect.ValueOf(obj)); value.IsValid() && value.Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
					seoField = value.Field(i).Interface().(Setting)
					break
				}
			}
		}
	}

	if seoField.EnabledCustomize {
		title = seoField.Title
		description = seoField.Description
		keywords = seoField.Keywords
	} else {
		seoSetting := collection.SettingResource.NewStruct().(QorSEOSettingInterface)
		if !db.Where("name = ?", name).First(seoSetting).RecordNotFound() {
			title = seoSetting.GetTitle()
			description = seoSetting.GetDescription()
			keywords = seoSetting.GetKeywords()
		}
	}

	siteWideSetting := collection.SettingResource.NewStruct()
	db.Where("is_global_seo = ? AND name = ?", true, collection.Name).First(siteWideSetting)
	tagValues := siteWideSetting.(QorSEOSettingInterface).GetGlobalSetting()

	if tagValues == nil {
		tagValues = map[string]string{}
	}

	if seo.Context != nil {
		for key, value := range seo.Context(objects...) {
			tagValues[key] = value
		}
	}

	title = replaceTags(title, seo.Varibles, tagValues)
	description = replaceTags(description, seo.Varibles, tagValues)
	keywords = replaceTags(keywords, seo.Varibles, tagValues)
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">\n<meta name=\"keywords\" content=\"%s\"/>", title, description, keywords))
}

// GetSeo get a Seo by name
func (collection *Collection) GetSEO(name string) *SEO {
	for _, s := range collection.registeredSEO {
		if s.Name == name {
			return s
		}
	}

	return &SEO{Name: name, collection: collection}
}

// SEOSettingURL get setting inline edit url by name
func (collection *Collection) SEOSettingURL(name string) string {
	qorAdmin := collection.resource.GetAdmin()
	return fmt.Sprintf("%v/%v/!seo_setting?name=%v", qorAdmin.GetRouter().Prefix, collection.resource.ToParam(), url.QueryEscape(name))
}

// ConfigureQorResource configure seoCollection for qor admin
func (collection *Collection) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		Admin := res.GetAdmin()
		collection.resource = res
		if collection.SettingResource == nil {
			collection.SettingResource = Admin.AddResource(&QorSEOSetting{}, &admin.Config{Invisible: true})
		}

		collection.SettingResource.UseTheme("seo")
		collection.SettingResource.EditAttrs("Name", "Setting")
		if nameMeta := collection.SettingResource.GetMeta("Name"); nameMeta != nil {
			nameMeta.Type = "hidden"
		}

		globalSettingRes := Admin.AddResource(collection.globalSetting, &admin.Config{Invisible: true})
		collection.globalResource = globalSettingRes

		res.Config.Singleton = true
		res.UseTheme("seo")

		router := Admin.GetRouter()
		controller := seoController{Collection: collection}
		router.Get(res.ToParam(), controller.Index)
		router.Put(fmt.Sprintf("%v/!seo_setting", res.ToParam()), controller.Update)
		router.Get(fmt.Sprintf("%v/!seo_setting", res.ToParam()), controller.InlineEdit)

		registerFuncMap(Admin)
	}
}

// Helpers
func replaceTags(originalVal string, validTags []string, values map[string]string) string {
	re := regexp.MustCompile("{{([a-zA-Z0-9]*)}}")
	matches := re.FindAllStringSubmatch(originalVal, -1)
	for _, match := range matches {
		originalVal = strings.Replace(originalVal, match[0], values[match[1]], 1)
	}
	return originalVal
}
