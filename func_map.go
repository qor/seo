package seo

import (
	"fmt"
	"reflect"
	"text/template"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

func (collection *Collection) seoSections(db *gorm.DB) func() []interface{} {
	return func() []interface{} {
		settings := []interface{}{}
		for _, seo := range collection.registeredSeo {
			s := collection.SettingResource.NewStruct()
			db.Where("name = ?", seo.Name).First(s)
			if db.NewRecord(s) {
				s.(QorSeoSettingInterface).SetName(seo.Name)
				s.(QorSeoSettingInterface).SetSeoType(seo.Name)
				db.Save(s)
			}
			settings = append(settings, s)
		}
		return settings
	}
}

func (collection *Collection) seoSettingMetas() []*admin.Section {
	return collection.SettingResource.NewAttrs("ID", "Name", "Setting")
}

func (collection *Collection) seoGlobalSetting(db *gorm.DB) func() interface{} {
	return func() interface{} {
		s := collection.SettingResource.NewStruct()
		db.Where("is_global = ?", true).First(s)
		if db.NewRecord(s) {
			s.(QorSeoSettingInterface).SetName("QorSeoGlobalSettings")
			s.(QorSeoSettingInterface).SetSeoType("QorSeoGlobalSettings")
			s.(QorSeoSettingInterface).SetIsGlobal(true)
			db.Save(s)
		}
		return s
	}
}

func (collection *Collection) seoGlobalSettingValue(setting map[string]string) interface{} {
	value := reflect.Indirect(reflect.ValueOf(collection.globalSetting))
	for i := 0; i < value.NumField(); i++ {
		fieldName := value.Type().Field(i).Name
		if setting[fieldName] != "" {
			value.Field(i).SetString(setting[fieldName])
		}
	}
	return value.Interface()
}

func (collection *Collection) seoGlobalSettingMetas(globalSettingRes *admin.Resource) func() []*admin.Section {
	return func() []*admin.Section {
		return globalSettingRes.NewAttrs()
	}
}

func (collection *Collection) seoTagsByType(name string) (tags []string) {
	seo := collection.GetSeo(name)
	if seo == nil {
		return []string{}
	}
	value := reflect.Indirect(reflect.ValueOf(collection.globalSetting))
	for i := 0; i < value.NumField(); i++ {
		tags = append(tags, value.Type().Field(i).Name)
	}
	for _, s := range collection.GetSeo(name).Varibles {
		tags = append(tags, s)
	}
	return tags
}

func (collection *Collection) seoAppendDefaultValue(db *gorm.DB) func(seoName string, resourceSeoValue interface{}) interface{} {
	return func(seoName string, resourceSeoValue interface{}) interface{} {
		globalInteface := collection.SettingResource.NewStruct()
		db.Where("name = ?", seoName).Find(globalInteface)
		globalSetting := globalInteface.(QorSeoSettingInterface)
		setting := resourceSeoValue.(Setting)
		if !setting.EnabledCustomize && setting.Title == "" && setting.Description == "" && setting.Keywords == "" {
			setting.Title = globalSetting.GetTitle()
			setting.Description = globalSetting.GetDescription()
			setting.Keywords = globalSetting.GetKeywords()
		}
		return setting
	}
}

func (collection *Collection) seoURLFor(db *gorm.DB, a *admin.Admin, res *admin.Resource) func(value interface{}) string {
	return func(value interface{}) string {
		return fmt.Sprintf("%v/%v/%v", a.GetRouter().Prefix, res.ToParam(), db.NewScope(value).PrimaryKeyValue())
	}
}

func (collection *Collection) registerFuncMap(db *gorm.DB, a *admin.Admin, res *admin.Resource, globalSettingRes *admin.Resource) {
	funcMaps := template.FuncMap{
		"seo_sections":             collection.seoSections(db),
		"seo_setting_metas":        collection.seoSettingMetas,
		"seo_global_setting_value": collection.seoGlobalSettingValue,
		"seo_global_setting_metas": collection.seoGlobalSettingMetas(globalSettingRes),
		"seo_global_setting":       collection.seoGlobalSetting(db),
		"seo_tags_by_type":         collection.seoTagsByType,
		"seo_append_default_value": collection.seoAppendDefaultValue(db),
		"seo_url_for":              collection.seoURLFor(db, a, res),
	}

	for key, value := range funcMaps {
		a.RegisterFuncMap(key, value)
	}
}
