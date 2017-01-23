package seo

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/qor/admin"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

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
