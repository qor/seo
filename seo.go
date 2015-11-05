package seo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"html/template"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"
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

var injected bool

func (setting Setting) Render(mainObj interface{}, obj interface{}) template.HTML {
	title := replaceTags(setting.Title, splitTags(setting.Tags), mainObj, obj)
	description := replaceTags(setting.Description, splitTags(setting.Tags), mainObj, obj)
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">", title, description))
}

// Configure
func (Setting) ConfigureQorResource(res *admin.Resource) {
	Admin := res.GetAdmin()
	scope := Admin.Config.DB.NewScope(res.Value)

	if !injected {
		injected = true
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/seo/views"))
		}
		res.UseTheme("seo")
	}

	for _, field := range scope.Fields() {
		if field.Struct.Type == reflect.TypeOf(Setting{}) {
			name := field.Name

			meta := res.GetMeta(name)
			if meta != nil {
				meta.Type = "seo"
			} else {
				res.Meta(&admin.Meta{Name: name, Type: "seo"})
				meta = res.GetMeta(name)
			}

			tags := field.Tag.Get("seo")
			tagsArray := splitTags(tags)
			meta.SetValuer(func(value interface{}, ctx *qor.Context) interface{} {
				settingField, _ := ctx.GetDB().NewScope(value).FieldByName(name)
				setting := settingField.Field.Interface().(settingInterface).GetSetting()
				setting.Tags = tags
				setting.TagsArray = tagsArray
				return setting
			})
		}
	}
	registerFunctions(res)
}

func registerFunctions(res *admin.Resource) {
	res.GetAdmin().RegisterFuncMap("filter_default_var_metas", func(metas []*admin.Meta) []*admin.Meta {
		var filterDefaultVarMetas []*admin.Meta
		for _, meta := range metas {
			if meta.Type != "seo" {
				filterDefaultVarMetas = append(filterDefaultVarMetas, meta)
			}
		}
		return filterDefaultVarMetas
	})

	res.GetAdmin().RegisterFuncMap("filter_page_metas", func(metas []*admin.Meta) []*admin.Meta {
		var filterPageMetas []*admin.Meta
		for _, meta := range metas {
			if meta.Type == "seo" {
				filterPageMetas = append(filterPageMetas, meta)
			}
		}
		return filterPageMetas
	})
}

// Helpers
func replaceTags(originalVal string, validTags []string, mainObj interface{}, obj interface{}) string {
	re := regexp.MustCompile("{{([a-zA-Z0-9]*)}}")
	matches := re.FindAllStringSubmatch(originalVal, -1)
	for _, match := range matches {
		field := reflect.ValueOf(obj).FieldByName(match[1])
		if field.IsValid() && isTagContains(validTags, match[1]) {
			value := field.Interface().(string)
			originalVal = strings.Replace(originalVal, match[0], value, 1)
		}
	}
	for _, match := range matches {
		field := reflect.ValueOf(mainObj).FieldByName(match[1])
		if field.IsValid() {
			value := field.Interface().(string)
			originalVal = strings.Replace(originalVal, match[0], value, 1)
		}
	}
	return originalVal
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
