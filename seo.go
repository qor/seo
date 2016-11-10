package seo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

type SeoCollection struct {
	SettingResource *admin.Resource
	registeredSeo   []*Seo
	setting         interface{}
}

type Seo struct {
	Name     string
	Settings []string
	Context  func(...interface{}) map[string]string
}

func init() {
	admin.RegisterViewPath("github.com/qor/seo/views")
}

func New() *SeoCollection {
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
	Name          string
	Setting       Setting `gorm:"size:4294967295"`
	seoCollection *SeoCollection
}

type QorSeoResourceInterface interface {
	GetSeoSetting() *Setting
}

type QorSeoSettingInterface interface {
	GetName() string
	SetName(name string)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
	SetSeoType(t string)
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

func (s *QorSeoSetting) SetSeoType(t string) {
	s.Setting.Type = t
}

func (s QorSeoSetting) GetSeoSetting() *Setting {
	return &s.Setting
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
					s.(QorSeoSettingInterface).SetSeoType(seo.Name)
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
				s.(QorSeoSettingInterface).SetSeoType("QorSeoGlobalSettings")
				db.Save(s)
			}
			return s
		})
		Admin.RegisterFuncMap("seoTagsByType", func(name string) []string {
			seo := seoCollection.GetSeo(name)
			if seo == nil {
				return []string{}
			}
			return seoCollection.GetSeo(name).Settings
		})
		Admin.RegisterFuncMap("append_default_seo_value", func(name string, value interface{}) interface{} {
			globalInteface := seoCollection.SettingResource.NewStruct()
			db.Where("name = ?", name).Find(globalInteface)
			globalSetting := globalInteface.(QorSeoSettingInterface)
			setting := value.(Setting)
			if !setting.EnabledCustomize && setting.Title == "" && setting.Description == "" && setting.Keywords == "" {
				setting.Title = globalSetting.GetTitle()
				setting.Description = globalSetting.GetDescription()
				setting.Keywords = globalSetting.GetKeywords()
			}
			return setting
		})

	}
}

// Render render SEO Setting
func (seoCollection SeoCollection) Render(name string, objects ...interface{}) template.HTML {
	var (
		title       string
		description string
		keywords    string
		setting     *Setting
	)
	for _, obj := range objects {
		value := reflect.ValueOf(obj)
		for i := 0; i < value.NumField(); i++ {
			if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
				s := value.Field(i).Interface().(Setting)
				setting = &s
				break
			}
		}
	}

	globalSetting := seoCollection.SettingResource.NewStruct()
	seoCollection.SettingResource.GetAdmin().Config.DB.Where("name = ?", name).Find(globalSetting)
	seo := seoCollection.GetSeo(name)
	if setting != nil && setting.EnabledCustomize {
		title = setting.Title
		description = setting.Description
		keywords = setting.Keywords
	} else {
		title = globalSetting.(QorSeoSettingInterface).GetTitle()
		description = globalSetting.(QorSeoSettingInterface).GetDescription()
		keywords = globalSetting.(QorSeoSettingInterface).GetKeywords()
	}
	title = replaceTags(title, seo.Settings, seo.Context(objects...))
	description = replaceTags(description, seo.Settings, seo.Context(objects...))
	keywords = replaceTags(keywords, seo.Settings, seo.Context(objects...))
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">\n<meta name=\"keywords\" content=\"%s\"/>", title, description, keywords))
}

func (seoCollection *SeoCollection) GetSeo(name string) *Seo {
	for _, s := range seoCollection.registeredSeo {
		if s.Name == name {
			return s
		}
	}
	return nil
}

// Setting could be used to field type for SEO Settings
type Setting struct {
	Title            string
	Description      string
	Keywords         string
	Tags             string
	TagsArray        []string `json:"-"`
	Type             string
	EnabledCustomize bool
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

// ConfigureQorMetaBeforeInitialize configure SEO setting for qor admin
func (Setting) ConfigureQorMetaBeforeInitialize(meta resource.Metaor) {
	if meta, ok := meta.(*admin.Meta); ok {
		meta.Type = "seo"
		res := meta.GetBaseResource().(*admin.Resource)
		res.GetAdmin().RegisterViewPath("github.com/qor/seo/views")
		res.GetAdmin().RegisterFuncMap("seoType", func(value interface{}, meta admin.Meta) string {
			typeFromTag := meta.FieldStruct.Struct.Tag.Get("seo")
			typeFromTag = utils.ParseTagOption(typeFromTag)["TYPE"]
			if typeFromTag != "" {
				return typeFromTag
			}
			return value.(Setting).Type
		})
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
func replaceTags(originalVal string, validTags []string, values map[string]string) string {
	re := regexp.MustCompile("{{([a-zA-Z0-9]*)}}")
	matches := re.FindAllStringSubmatch(originalVal, -1)
	for _, match := range matches {
		originalVal = strings.Replace(originalVal, match[0], values[match[1]], 1)
	}
	return originalVal
}
