package seo

import (
	"fmt"
	"reflect"
	"text/template"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

func (seoCollection *SeoCollection) seoSections(db *gorm.DB) func() []interface{} {
	return func() []interface{} {
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
	}
}

func (seoCollection *SeoCollection) seoSettingMetas() []*admin.Section {
	return seoCollection.SettingResource.NewAttrs("ID", "Name", "Setting")
}

func (seoCollection *SeoCollection) seoGlobalSetting(db *gorm.DB) func() interface{} {
	return func() interface{} {
		s := seoCollection.SettingResource.NewStruct()
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

func (seoCollection *SeoCollection) seoGlobalSettingValue(setting map[string]string) interface{} {
	value := reflect.Indirect(reflect.ValueOf(seoCollection.globalSetting))
	for i := 0; i < value.NumField(); i++ {
		fieldName := value.Type().Field(i).Name
		if setting[fieldName] != "" {
			value.Field(i).SetString(setting[fieldName])
		}
	}
	return value.Interface()
}

func (seoCollection *SeoCollection) seoGlobalSettingMetas(globalSettingRes *admin.Resource) func() []*admin.Section {
	return func() []*admin.Section {
		return globalSettingRes.NewAttrs()
	}
}

func (seoCollection *SeoCollection) seoTagsByType(name string) (tags []string) {
	seo := seoCollection.GetSeo(name)
	if seo == nil {
		return []string{}
	}
	value := reflect.Indirect(reflect.ValueOf(seoCollection.globalSetting))
	for i := 0; i < value.NumField(); i++ {
		tags = append(tags, value.Type().Field(i).Name)
	}
	for _, s := range seoCollection.GetSeo(name).Settings {
		tags = append(tags, s)
	}
	return tags
}

func (seoCollection *SeoCollection) seoAppendDefaultValue(db *gorm.DB) func(seoName string, resourceSeoValue interface{}) interface{} {
	return func(seoName string, resourceSeoValue interface{}) interface{} {
		globalInteface := seoCollection.SettingResource.NewStruct()
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

func (seoCollection *SeoCollection) seoURLFor(db *gorm.DB, a *admin.Admin, res *admin.Resource) func(value interface{}) string {
	return func(value interface{}) string {
		return fmt.Sprintf("%v/%v/%v", a.GetRouter().Prefix, res.ToParam(), db.NewScope(value).PrimaryKeyValue())
	}
}

func (seoCollection *SeoCollection) registerFuncMap(db *gorm.DB, a *admin.Admin, res *admin.Resource, globalSettingRes *admin.Resource) {
	funcMaps := template.FuncMap{
		"seo_sections":             seoCollection.seoSections(db),
		"seo_setting_metas":        seoCollection.seoSettingMetas,
		"seo_global_setting_value": seoCollection.seoGlobalSettingValue,
		"seo_global_setting_metas": seoCollection.seoGlobalSettingMetas(globalSettingRes),
		"seo_global_setting":       seoCollection.seoGlobalSetting(db),
		"seo_tags_by_type":         seoCollection.seoTagsByType,
		"seo_append_default_value": seoCollection.seoAppendDefaultValue(db),
		"seo_url_for":              seoCollection.seoURLFor(db, a, res),
	}

	for key, value := range funcMaps {
		a.RegisterFuncMap(key, value)
	}
}
