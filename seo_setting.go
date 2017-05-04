package seo

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/qor/admin"
	"github.com/qor/qor/resource"
)

// QorSEOSettingInterface support customize Seo model
type QorSEOSettingInterface interface {
	GetName() string
	SetName(string)
	GetGlobalSetting() map[string]string
	SetGlobalSetting(map[string]string)
	GetSEOType() string
	SetSEOType(string)
	GetIsGlobalSEO() bool
	SetIsGlobalSEO(bool)
	GetTitle() string
	GetDescription() string
	GetKeywords() string
	SetCollection(*Collection)
}

// QorSEOSetting default seo model
type QorSEOSetting struct {
	Name        string `gorm:"primary_key"`
	Setting     Setting
	IsGlobalSEO bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	collection *Collection
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
func (s QorSEOSetting) GetName() string {
	return s.Name
}

// SetName set QorSeoSetting's name
func (s *QorSEOSetting) SetName(name string) {
	s.Name = name
}

// GetSeoType get QorSeoSetting's type
func (s QorSEOSetting) GetSEOType() string {
	return s.Setting.Type
}

// SetSeoType set QorSeoSetting's type
func (s *QorSEOSetting) SetSEOType(t string) {
	s.Setting.Type = t
}

// GetIsGlobalSEO get QorSEOSetting's isGlobal
func (s QorSEOSetting) GetIsGlobalSEO() bool {
	return s.IsGlobalSEO
}

// SetIsGlobalSEO set QorSeoSetting's isGlobal
func (s *QorSEOSetting) SetIsGlobalSEO(isGlobal bool) {
	s.IsGlobalSEO = isGlobal
}

// GetGlobalSetting get QorSeoSetting's globalSetting
func (s QorSEOSetting) GetGlobalSetting() map[string]string {
	return s.Setting.GlobalSetting
}

// SetGlobalSetting set QorSeoSetting's globalSetting
func (s *QorSEOSetting) SetGlobalSetting(globalSetting map[string]string) {
	s.Setting.GlobalSetting = globalSetting
}

// GetTitle get Setting's title
func (s QorSEOSetting) GetTitle() string {
	return s.Setting.Title
}

// GetDescription get Setting's description
func (s QorSEOSetting) GetDescription() string {
	return s.Setting.Description
}

// GetKeywords get Setting's keywords
func (s QorSEOSetting) GetKeywords() string {
	return s.Setting.Keywords
}

// SetCollection set Setting's collection
func (s *QorSEOSetting) SetCollection(collection *Collection) {
	s.collection = collection
}

// GetSEO get Setting's SEO configure
func (s QorSEOSetting) GetSEO() *SEO {
	return s.collection.GetSEO(s.Name)
}

// Scan scan value from database into struct
func (setting *Setting) Scan(value interface{}) error {
	if bytes, ok := value.([]byte); ok {
		json.Unmarshal(bytes, setting)
	} else if str, ok := value.(string); ok {
		json.Unmarshal([]byte(str), setting)
	} else if strs, ok := value.([]string); ok {
		for _, str := range strs {
			json.Unmarshal([]byte(str), setting)
		}
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
		if res, ok := meta.GetBaseResource().(*admin.Resource); ok {
			res.UseTheme("seo_meta")
		}
	}
}
