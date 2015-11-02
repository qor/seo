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
	"strings"
)

type Setting struct {
	Title       string
	Description string
	Vars        []string
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

			var vars []string
			tags := strings.Split(field.Tag.Get("seo"), ",")
			for _, tag := range tags {
				vars = append(vars, strings.Trim(tag, " "))
			}
			meta.Valuer = func(value interface{}, ctx *qor.Context) interface{} {
				settingField, _ := ctx.GetDB().NewScope(value).FieldByName(name)
				setting := settingField.Field.Interface().(settingInterface).GetSetting()
				setting.Vars = vars
				return setting
			}
		}
	}
}

func (setting Setting) Render() template.HTML {
	title := setting.Title
	description := setting.Description
	return template.HTML(fmt.Sprintf("<title>%s</title>\n<meta name=\"description\" content=\"%s\">", title, description))
}
