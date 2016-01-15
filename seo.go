package seo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

type Setting struct {
	Title       string
	Description string
	Tags        string
	TagsArray   []string `json:"-"`
}

type settingInterface interface {
	GetSetting() Setting
}

func (setting Setting) GetSetting() Setting {
	return setting
}

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

func (setting Setting) Value() (driver.Value, error) {
	result, err := json.Marshal(setting)
	return string(result), err
}

func (setting Setting) Render(mainObj interface{}, obj interface{}) template.HTML {
	objTags := splitTags(setting.Tags)
	reflectValue := reflect.Indirect(reflect.ValueOf(mainObj))
	allTags := prependMainObjectTags(objTags, reflectValue)
	title := replaceTags(setting.Title, allTags, mainObj, obj)
	description := replaceTags(setting.Description, allTags, mainObj, obj)
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">", title, description))
}

// Configure
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
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/seo/views"))
		}
		res.UseTheme("seo")
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
func replaceTags(originalVal string, validTags []string, mainObj interface{}, obj interface{}) string {
	re := regexp.MustCompile("{{([a-zA-Z0-9]*)}}")
	matches := re.FindAllStringSubmatch(originalVal, -1)
	originalVal = replaceValues(matches, obj, originalVal)
	return replaceValues(matches, mainObj, originalVal)
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

func replaceValues(matches [][]string, obj interface{}, originalVal string) string {
	if obj == nil {
		return originalVal
	}
	for _, match := range matches {
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
	return originalVal
}
