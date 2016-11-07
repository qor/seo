package seo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type SeoCollection struct {
	SettingResource *admin.Resource
	registeredSeo   []*Seo
	setting         interface{}
}

type Seo struct {
	Name     string
	Settings []string
	Context  func(...interface{}) map[string]interface{}
}

func init() {
	admin.RegisterViewPath("github.com/qor/seo/views")
}

func New() *SeoCollection {
	/*settingRes := a.AddResource(&QorSeoSetting{}, &admin.Config{Invisible: false})
	settingRes.Meta(&admin.Meta{Name: "Name", Type: "hidden"})
	res.UseTheme("seo")*/
	return &SeoCollection{}
}

func (seoCollection *SeoCollection) RegisterGlobalSetting(s interface{}) {
	seoCollection.setting = s
}

func (seoCollection *SeoCollection) RegisterSeo(seo *Seo) {
	seoCollection.registeredSeo = append(seoCollection.registeredSeo, seo)
}

type QorSeoSetting struct {
	gorm.Model
	Name    string
	Setting Setting `gorm:"size:4294967295"`
}

type QorSeoSettingInterface interface {
	GetName() string
	SetName(name string)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
}

func (s QorSeoSetting) GetName() string {
	return s.Name
}

func (s *QorSeoSetting) SetName(name string) {
	s.Name = name
}

func (s QorSeoSetting) GetTitle() string {
	return s.Setting.Title
}

func (s QorSeoSetting) GetDescription() string {
	return s.Setting.Description
}

func (s QorSeoSetting) GetKeywords() string {
	return s.Setting.Keywords
}

func (seoCollection *SeoCollection) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		Admin := res.GetAdmin()
		db := Admin.Config.DB
		if seoCollection.SettingResource == nil {
			seoCollection.SettingResource = res.GetAdmin().AddResource(&QorSeoSetting{}, &admin.Config{Invisible: true})
		}
		seoCollection.SettingResource.UseTheme("seo")
		nameMeta := seoCollection.SettingResource.GetMeta("Name")
		nameMeta.Type = "hidden"

		res.Config.Singleton = true
		res.UseTheme("seo")
		router := Admin.GetRouter()
		controller := seoController{SeoCollection: seoCollection}
		router.Get(res.ToParam(), controller.Index)
		Admin.RegisterFuncMap("seoSections", func() []interface{} {
			settings := []interface{}{}
			for _, seo := range seoCollection.registeredSeo {
				s := seoCollection.SettingResource.NewStruct()
				db.Where("name = ?", seo.Name).First(s)
				if db.NewRecord(s) {
					s.(QorSeoSettingInterface).SetName(seo.Name)
					db.Save(s)
				}
				settings = append(settings, s)
			}
			return settings
		})
		Admin.RegisterFuncMap("globalSeoSection", func() interface{} {
			s := seoCollection.SettingResource.NewStruct()
			db.Where("name = ?", "QorSeoGlobalSettings").First(s)
			if db.NewRecord(s) {
				s.(QorSeoSettingInterface).SetName("QorSeoGlobalSettings")
				db.Save(s)
			}
			return s
		})
	}
}

// Render render SEO Setting
func (seoCollection SeoCollection) Render(name string, objects ...interface{}) template.HTML {
	seoSetting := seoCollection.SettingResource.NewStruct()
	seoCollection.SettingResource.GetAdmin().Config.DB.Where("name = ?", name).Find(seoSetting)
	seo := seoCollection.getSeo(name)
	_ = seo.Context(objects...)
	/*objTags := splitTags(seoCollection.Setting.Tags)
	reflectValue := reflect.Indirect(reflect.ValueOf(seoSetting))
	allTags := prependMainObjectTags(objTags, reflectValue)
	title := replaceTags(setting.Title, allTags, seoSetting, obj...)
	description := replaceTags(setting.Description, allTags, seoSetting, obj...)
	keywords := replaceTags(setting.Keywords, allTags, seoSetting, obj...)*/
	title := seoSetting.(QorSeoSettingInterface).GetTitle()
	description := seoSetting.(QorSeoSettingInterface).GetDescription()
	keywords := seoSetting.(QorSeoSettingInterface).GetKeywords()
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">\n<meta name=\"keywords\" content=\"%s\"/>", title, description, keywords))
}

func (seoCollection *SeoCollection) getSeo(name string) *Seo {
	for _, s := range seoCollection.registeredSeo {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Setting could be used to field type for SEO Settings
type Setting struct {
	Title       string
	Description string
	Keywords    string
	Tags        string
	TagsArray   []string `json:"-"`
}

type settingInterface interface {
	GetSetting() Setting
}

// GetSetting return itself to match interface
func (setting Setting) GetSetting() Setting {
	return setting
}

// Scan scan value from database into struct
func (setting *Setting) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		json.Unmarshal(bytes, setting)
	} else if str, ok := value.(string); ok {
		json.Unmarshal([]byte(str), setting)
	} else if strs, ok := value.([]string); ok {
		json.Unmarshal([]byte(strs[0]), setting)
	}
	return nil
}

// Value get value from struct, and save into database
func (setting Setting) Value() (driver.Value, error) {
	result, err := json.Marshal(setting)
	return string(result), err
}

// Render render SEO Setting
func (setting Setting) Render(seoSetting interface{}, obj ...interface{}) template.HTML {
	objTags := splitTags(setting.Tags)
	reflectValue := reflect.Indirect(reflect.ValueOf(seoSetting))
	allTags := prependMainObjectTags(objTags, reflectValue)
	title := replaceTags(setting.Title, allTags, seoSetting, obj...)
	description := replaceTags(setting.Description, allTags, seoSetting, obj...)
	keywords := replaceTags(setting.Keywords, allTags, seoSetting, obj...)
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">\n<meta name=\"keywords\" content=\"%s\"/>", title, description, keywords))
}

// ConfigureQorMetaBeforeInitialize configure SEO setting for qor admin
func (Setting) ConfigureQorMetaBeforeInitialize(meta resource.Metaor) {
	if meta, ok := meta.(*admin.Meta); ok {
		meta.Type = "seo"

		if meta.GetValuer() == nil {
			res := meta.GetBaseResource().(*admin.Resource)
			Admin := res.GetAdmin()

			tags := meta.FieldStruct.Struct.Tag.Get("seo")
			tagsArray := splitTags(tags)
			tagsArray = prependMainObjectTags(tagsArray, Admin.Config.DB.NewScope(res.Value).IndirectValue())

			meta.SetValuer(func(value interface{}, ctx *qor.Context) interface{} {
				settingField, _ := ctx.GetDB().NewScope(value).FieldByName(meta.FieldStruct.Struct.Name)
				setting := settingField.Field.Interface().(settingInterface).GetSetting()
				setting.Tags = tags
				setting.TagsArray = tagsArray
				return setting
			})
		}

		res := meta.GetBaseResource().(*admin.Resource)
		res.GetAdmin().RegisterViewPath("github.com/qor/seo/views")
		//res.UseTheme("seo")
		registerFunctions(res)
	}
}

func registerFunctions(res *admin.Resource) {
	res.GetAdmin().RegisterFuncMap("filter_default_var_sections", func(sections []*admin.Section) []*admin.Section {
		var filterDefaultVarSections []*admin.Section
		for _, section := range sections {
			isContainSeoTag := false
			for _, row := range section.Rows {
				for _, col := range row {
					meta := res.GetMetaOrNew(col)
					if meta != nil && meta.Type == "seo" {
						isContainSeoTag = true
					}
				}
			}
			if !isContainSeoTag {
				filterDefaultVarSections = append(filterDefaultVarSections, section)
			}
		}
		return filterDefaultVarSections
	})

	res.GetAdmin().RegisterFuncMap("filter_page_sections", func(sections []*admin.Section) []*admin.Section {
		var filterPageSections []*admin.Section
		for _, section := range sections {
			isContainSeoTag := false
			for _, row := range section.Rows {
				for _, col := range row {
					meta := res.GetMetaOrNew(col)
					if meta != nil && meta.Type == "seo" {
						isContainSeoTag = true
					}
				}
			}
			if isContainSeoTag {
				filterPageSections = append(filterPageSections, section)
			}
		}
		return filterPageSections
	})
}

// Helpers
func replaceTags(originalVal string, validTags []string, mainObj interface{}, obj ...interface{}) string {
	re := regexp.MustCompile("{{([a-zA-Z0-9]*)}}")
	matches := re.FindAllStringSubmatch(originalVal, -1)
	return replaceValues(originalVal, matches, append(obj, mainObj)...)
}

func isTagContains(tags []string, item string) bool {
	for _, t := range tags {
		if item == t {
			return true
		}
	}
	return false
}

func splitTags(tags string) []string {
	var tagsArray []string
	for _, tag := range strings.Split(tags, ",") {
		tagsArray = append(tagsArray, strings.Trim(tag, " "))
	}
	return tagsArray
}

func prependMainObjectTags(tags []string, mainValue reflect.Value) []string {
	var results []string
	if mainValue.Kind() == reflect.Struct {
		for i := 0; i < mainValue.NumField(); i++ {
			if mainValue.Field(i).Kind() == reflect.String {
				results = append(results, mainValue.Type().Field(i).Name)
			}
		}
	}
	for _, tag := range tags {
		if tag != "" {
			results = append(results, tag)
		}
	}
	return results
}

func replaceValues(originalVal string, matches [][]string, objs ...interface{}) string {
	for _, match := range matches {
		for _, obj := range objs {
			reflectValue := reflect.Indirect(reflect.ValueOf(obj))
			if reflectValue.Kind() == reflect.Struct {
				field := reflectValue.FieldByName(match[1])
				if field.IsValid() {
					value := field.Interface().(string)
					originalVal = strings.Replace(originalVal, match[0], value, 1)
				}
			} else {
				color.Yellow("[WARNING] Qor SEO: The parameter you passed is not a Struct")
			}
		}
	}
	return originalVal
}
