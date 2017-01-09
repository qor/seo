package seo

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/qor/admin"
	"github.com/qor/responder"
)

type seoController struct {
	Collection   *Collection
	MainResource *admin.Resource
}

func (sc seoController) Index(context *admin.Context) {
	context = context.NewResourceContext(sc.Collection.SettingResource)
	context.Execute("index", struct {
		Collection      *Collection
		SettingResource *admin.Resource
	}{
		Collection:      sc.Collection,
		SettingResource: sc.Collection.SettingResource,
	})
}

func (sc seoController) InlineEdit(context *admin.Context) {
	context = context.NewResourceContext(sc.Collection.SettingResource)
	result := sc.Collection.SettingResource.NewStruct()
	name, err := url.QueryUnescape(context.Request.Form.Get("name"))
	if err != nil {
		context.AddError(err)
	}
	context.DB.Where("name = ?", name).First(result)

	result.(QorSeoSettingInterface).SetCollection(sc.Collection)
	responder.With("html", func() {
		context.Execute("edit", struct {
			Setting interface{}
			EditURL string
			Metas   []*admin.Section
		}{
			Setting: result,
			EditURL: sc.Collection.SeoSettingURL(name),
			Metas:   seoSettingMetas(sc.Collection),
		})
	}).With("json", func() {
		context.JSON("edit", result)
	}).Respond(context.Request)
}

func (sc seoController) Update(context *admin.Context) {
	context = context.NewResourceContext(sc.Collection.SettingResource)
	result := sc.Collection.SettingResource.NewStruct()
	name, err := url.QueryUnescape(context.Request.Form.Get("name"))
	if err != nil {
		context.AddError(err)
	}
	context.DB.Where("name = ?", name).First(result)
	if context.DB.NewRecord(result) {
		context.Request.Form["QorResource.Name"] = []string{name}
		context.Request.Form["QorResource.Setting.Type"] = []string{name}
	}

	seoSettingInterface := result.(QorSeoSettingInterface)
	if seoSettingInterface.GetIsGlobalSeo() {
		globalSetting := make(map[string]string)
		for fieldWithPrefix := range context.Request.Form {
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
