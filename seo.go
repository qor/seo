package seo

import (
	"github.com/qor/qor/admin"
	"os"
	"path"
	"reflect"
	"strings"
)

type Setting struct {
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

			res.Meta(&admin.Meta{Name: name, Type: "seo"})
		}
	}

}
