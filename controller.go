package seo

import (
	"net/http"
	"path"
	"strings"

	"github.com/qor/admin"
	"github.com/qor/responder"
)

type seoController struct {
	SeoCollection *SeoCollection
	MainResource  *admin.Resource
}

func (sc seoController) Index(context *admin.Context) {
	context = context.NewResourceContext(sc.SeoCollection.SettingResource)
	context.Execute("index", struct {
		SettingResource *admin.Resource
	}{
		SettingResource: sc.SeoCollection.SettingResource,
	})
}

func (sc seoController) Update(context *admin.Context) {
	context = context.NewResourceContext(sc.SeoCollection.SettingResource)
	var result interface{}
	var err error

	result, err = context.FindOne()
	context.AddError(err)

	seoSettingInterface := result.(QorSeoSettingInterface)
	if seoSettingInterface.GetName() == "QorSeoGlobalSettings" {
		globalSetting := make(map[string]string)
		for fieldWithPrefix, _ := range context.Request.Form {
			if strings.HasPrefix(fieldWithPrefix, "QorResource") {
				field := strings.Replace(fieldWithPrefix, "QorResource.", "", -1)
				globalSetting[field] = context.Request.Form.Get(fieldWithPrefix)
			}
		}
		seoSettingInterface.SetGlobalSetting(globalSetting)
	}

	res := context.Resource
	if !context.HasError() {
		if context.AddError(res.Decode(context.Context, result)); !context.HasError() {
			context.AddError(res.CallSave(result, context.Context))
		}
	}

	responder.With("html", func() {
		http.Redirect(context.Writer, context.Request, path.Join(sc.MainResource.GetAdmin().GetRouter().Prefix, sc.MainResource.ToParam()), http.StatusFound)
	}).With("json", func() {
		if context.HasError() {
			context.Writer.WriteHeader(admin.HTTPUnprocessableEntity)
			context.JSON("edit", map[string]interface{}{"errors": context.GetErrors()})
		} else {
			context.JSON("show", result)
		}
	}).Respond(context.Request)
}
