package seo

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"os"
	"path"
	"reflect"
	"strings"
)

type Setting struct {
	Title       string
	Description string
	Vars        string
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
	}
	return nil
}

func (setting Setting) Value() (driver.Value, error) {
	return setting.Title, nil
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

			tags := field.Tag.Get("seo")
			meta.Valuer = func(value interface{}, ctx *qor.Context) interface{} {
				settingField, _ := ctx.GetDB().NewScope(value).FieldByName(name)
				setting := settingField.Field.Interface().(settingInterface).GetSetting()
				setting.Vars = tags
				return setting
			}
		}
	}
}
