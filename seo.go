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
	globalSetting   interface{}
}

type Seo struct {
	Name     string
	Settings []string
	Context  func(...interface{}) map[string]string
}

type QorSeoSetting struct {
	gorm.Model
	Name          string
	Setting       Setting `gorm:"size:4294967295"`
	seoCollection *SeoCollection
	IsGlobal      bool
}

type QorSeoSettingInterface interface {
	GetName() string
	SetName(string)
	GetGlobalSetting() map[string]string
	SetGlobalSetting(map[string]string)
	GetSeoType() string
	SetSeoType(string)
	GetIsGlobal() bool
	SetIsGlobal(bool)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
}

// Setting could be used to field type for SEO Settings
type Setting struct {
	Title            string
	Description      string
	Keywords         string
	Type             string
	EnabledCustomize bool
	GlobalSetting    map[string]string
}

func init() {
	admin.RegisterViewPath("github.com/qor/seo/views")
}

func New() *SeoCollection {
	return &SeoCollection{}
}

func (seoCollection *SeoCollection) RegisterGlobalSetting(s interface{}) {
	seoCollection.globalSetting = s
}

func (seoCollection *SeoCollection) RegisterSeo(seo *Seo) {
	seoCollection.registeredSeo = append(seoCollection.registeredSeo, seo)
}

func (s QorSeoSetting) GetName() string {
	return s.Name
}

func (s *QorSeoSetting) SetName(name string) {
	s.Name = name
}

func (s QorSeoSetting) GetSeoType() string {
	return s.Setting.Type
}

func (s *QorSeoSetting) SetSeoType(t string) {
	s.Setting.Type = t
}

func (s QorSeoSetting) GetIsGlobal() bool {
	return s.IsGlobal
}

func (s *QorSeoSetting) SetIsGlobal(isGlobal bool) {
	s.IsGlobal = isGlobal
}

func (s QorSeoSetting) GetGlobalSetting() map[string]string {
	return s.Setting.GlobalSetting
}

func (s *QorSeoSetting) SetGlobalSetting(globalSetting map[string]string) {
	s.Setting.GlobalSetting = globalSetting
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
		nameMeta := seoCollection.SettingResource.GetMetaOrNew("Name")
		nameMeta.Type = "hidden"
		globalSettingRes := Admin.AddResource(seoCollection.globalSetting, &admin.Config{Invisible: true})

		res.Config.Singleton = true
		res.UseTheme("seo")
		router := Admin.GetRouter()
		controller := seoController{SeoCollection: seoCollection, MainResource: res}
		router.Get(res.ToParam(), controller.Index)
		router.Put(fmt.Sprintf("%v/%v", res.ToParam(), seoCollection.SettingResource.ParamIDName()), controller.Update)

		seoCollection.registerFuncMap(db, Admin, res, globalSettingRes)
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
	db := seoCollection.SettingResource.GetAdmin().Config.DB
	for _, obj := range objects {
		value := reflect.ValueOf(obj)
		if value.IsValid() && value.Kind().String() == "struct" {
			for i := 0; i < value.NumField(); i++ {
				if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
					s := value.Field(i).Interface().(Setting)
					setting = &s
					break
				}
			}
		}
	}

	globalSetting := seoCollection.SettingResource.NewStruct()
	db.Where("name = ?", name).Find(globalSetting)
	seo := seoCollection.GetSeo(name)
	if seo == nil {
		utils.ExitWithMsg(fmt.Printf("SEO: Can't find seo with name %v", name))
		return ""
	}

	if setting != nil && setting.EnabledCustomize {
		title = setting.Title
		description = setting.Description
		keywords = setting.Keywords
	} else {
		title = globalSetting.(QorSeoSettingInterface).GetTitle()
		description = globalSetting.(QorSeoSettingInterface).GetDescription()
		keywords = globalSetting.(QorSeoSettingInterface).GetKeywords()
	}

	var tagValues map[string]string
	if seo.Context != nil {
		tagValues = seo.Context(objects...)
	} else {
		tagValues = make(map[string]string)
	}
	s := seoCollection.SettingResource.NewStruct()
	db.Where("is_global = ?", true).First(s)
	for k, v := range s.(QorSeoSettingInterface).GetGlobalSetting() {
		if tagValues[k] == "" {
			tagValues[k] = v
		}
	}
	title = replaceTags(title, seo.Settings, tagValues)
	description = replaceTags(description, seo.Settings, tagValues)
	keywords = replaceTags(keywords, seo.Settings, tagValues)
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
		res.GetAdmin().RegisterFuncMap("seoType", func(value interface{}, meta admin.Meta) string {
			typeFromTag := meta.FieldStruct.Struct.Tag.Get("seo")
			typeFromTag = utils.ParseTagOption(typeFromTag)["TYPE"]
			if typeFromTag != "" {
				return typeFromTag
			}
			return value.(Setting).Type
		})
		res.UseTheme("seo_meta")
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
