package seo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

// Collection will hold registered seo configures and global setting definition and other configures
type Collection struct {
	SettingResource   *admin.Resource
	GlobalSettingName string
	registeredSeo     []*SEO
	globalSetting     interface{}
	resource          *admin.Resource
	globalResource    *admin.Resource
}

// SEO represents a seo object for a page
type SEO struct {
	Name       string
	Varibles   []string
	Context    func(...interface{}) map[string]string
	collection *Collection
}

// QorSeoSetting default seo model
type QorSeoSetting struct {
	Name        string `gorm:"primary_key"`
	Setting     Setting
	collection  *Collection
	IsGlobalSeo bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// Setting defined meta's attributes
type Setting struct {
	Title            string `gorm:"size:4294967295"`
	Description      string
	Keywords         string
	Type             string
	EnabledCustomize bool
	GlobalSetting    map[string]string
}

// QorSeoSettingInterface support customize Seo model
type QorSeoSettingInterface interface {
	GetName() string
	SetName(string)
	GetGlobalSetting() map[string]string
	SetGlobalSetting(map[string]string)
	GetSeoType() string
	SetSeoType(string)
	GetIsGlobalSeo() bool
	SetIsGlobalSeo(bool)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
	SetCollection(*Collection)
}

func init() {
	admin.RegisterViewPath("github.com/qor/seo/views")
}

// New initialize a SeoCollection instance
func New(name string) *Collection {
	return &Collection{GlobalSettingName: name}
}

// RegisterGlobalVaribles register global setting and will represents as 'Site-wide Settings' part in admin
func (collection *Collection) RegisterGlobalVaribles(s interface{}) {
	collection.globalSetting = s
}

// RegisterSeo register a seo
func (collection *Collection) RegisterSeo(seo *SEO) {
	seo.collection = collection
	collection.registeredSeo = append(collection.registeredSeo, seo)
}

// GetName get QorSeoSetting's name
func (s QorSeoSetting) GetName() string {
	return s.Name
}

// SetName set QorSeoSetting's name
func (s *QorSeoSetting) SetName(name string) {
	s.Name = name
}

// GetSeoType get QorSeoSetting's type
func (s QorSeoSetting) GetSeoType() string {
	return s.Setting.Type
}

// SetSeoType set QorSeoSetting's type
func (s *QorSeoSetting) SetSeoType(t string) {
	s.Setting.Type = t
}

// GetIsGlobalSeo get QorSeoSetting's isGlobal
func (s QorSeoSetting) GetIsGlobalSeo() bool {
	return s.IsGlobalSeo
}

// SetIsGlobalSeo set QorSeoSetting's isGlobal
func (s *QorSeoSetting) SetIsGlobalSeo(isGlobal bool) {
	s.IsGlobalSeo = isGlobal
}

// GetGlobalSetting get QorSeoSetting's globalSetting
func (s QorSeoSetting) GetGlobalSetting() map[string]string {
	return s.Setting.GlobalSetting
}

// SetGlobalSetting set QorSeoSetting's globalSetting
func (s *QorSeoSetting) SetGlobalSetting(globalSetting map[string]string) {
	s.Setting.GlobalSetting = globalSetting
}

// GetTitle get Setting's title
func (s QorSeoSetting) GetTitle() string {
	return s.Setting.Title
}

// GetDescription get Setting's description
func (s QorSeoSetting) GetDescription() string {
	return s.Setting.Description
}

// GetKeywords get Setting's keywords
func (s QorSeoSetting) GetKeywords() string {
	return s.Setting.Keywords
}

// SetCollection set Setting's collection
func (s *QorSeoSetting) SetCollection(collection *Collection) {
	s.collection = collection
}

// GetSeo get Setting's SEO configure
func (s QorSeoSetting) GetSeo() *SEO {
	return s.collection.GetSeo(s.Name)
}

// ConfigureQorResource configure seoCollection for qor admin
func (collection *Collection) ConfigureQorResource(res resource.Resourcer) {
	if res, ok := res.(*admin.Resource); ok {
		Admin := res.GetAdmin()
		collection.resource = res
		if collection.SettingResource == nil {
			collection.SettingResource = res.GetAdmin().AddResource(&QorSeoSetting{}, &admin.Config{Invisible: true})
		}
		collection.SettingResource.UseTheme("seo")
		collection.SettingResource.EditAttrs("Name", "Setting")
		nameMeta := collection.SettingResource.GetMetaOrNew("Name")
		nameMeta.Type = "hidden"
		globalSettingRes := Admin.AddResource(collection.globalSetting, &admin.Config{Invisible: true})
		collection.globalResource = globalSettingRes

		res.Config.Singleton = true
		res.UseTheme("seo")
		router := Admin.GetRouter()
		controller := seoController{Collection: collection, MainResource: res}
		router.Get(res.ToParam(), controller.Index)
		router.Put(fmt.Sprintf("%v/!seo_setting", res.ToParam()), controller.Update)
		router.Get(fmt.Sprintf("%v/!seo_setting", res.ToParam()), controller.InlineEdit)

		registerFuncMap(Admin)
	}
}

// Render render SEO Setting
func (collection Collection) Render(context *qor.Context, name string, objects ...interface{}) template.HTML {
	var (
		title           string
		description     string
		keywords        string
		resourceSetting *Setting
	)

	db := context.GetDB()
	for _, obj := range objects {
		value := reflect.ValueOf(obj)
		if value.IsValid() && value.Kind().String() == "struct" {
			for i := 0; i < value.NumField(); i++ {
				if value.Field(i).Type() == reflect.TypeOf(Setting{}) {
					s := value.Field(i).Interface().(Setting)
					resourceSetting = &s
					break
				}
			}
		}
	}

	shareSetting := collection.SettingResource.NewStruct()
	db.Where("name = ?", name).First(shareSetting)
	seo := collection.GetSeo(name)
	if seo == nil {
		utils.ExitWithMsg(fmt.Printf("SEO: Can't find seo with name %v", name))
		return ""
	}

	if resourceSetting != nil && resourceSetting.EnabledCustomize {
		title = resourceSetting.Title
		description = resourceSetting.Description
		keywords = resourceSetting.Keywords
	} else {
		title = shareSetting.(QorSeoSettingInterface).GetTitle()
		description = shareSetting.(QorSeoSettingInterface).GetDescription()
		keywords = shareSetting.(QorSeoSettingInterface).GetKeywords()
	}

	var tagValues map[string]string
	if seo.Context != nil {
		tagValues = seo.Context(objects...)
	} else {
		tagValues = make(map[string]string)
	}
	siteWideSetting := collection.SettingResource.NewStruct()
	db.Where("is_global_seo = ? AND name = ?", true, collection.GlobalSettingName).First(siteWideSetting)
	for k, v := range siteWideSetting.(QorSeoSettingInterface).GetGlobalSetting() {
		if tagValues[k] == "" {
			tagValues[k] = v
		}
	}
	title = replaceTags(title, seo.Varibles, tagValues)
	description = replaceTags(description, seo.Varibles, tagValues)
	keywords = replaceTags(keywords, seo.Varibles, tagValues)
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">\n<meta name=\"keywords\" content=\"%s\"/>", title, description, keywords))
}

// GetSeo get a Seo by name
func (collection *Collection) GetSeo(name string) *SEO {
	for _, s := range collection.registeredSeo {
		if s.Name == name {
			return s
		}
	}
	newSeo := &SEO{Name: name}
	newSeo.collection = collection
	return newSeo
}

// SeoSettingURL get setting inline edit url by name
func (collection *Collection) SeoSettingURL(name string) string {
	qorAdmin := collection.resource.GetAdmin()
	return fmt.Sprintf("%v/%v/!seo_setting?name=%v", qorAdmin.GetRouter().Prefix, collection.resource.ToParam(), url.QueryEscape(name))
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
