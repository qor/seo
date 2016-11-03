package seo

import "github.com/qor/admin"

type seoController struct {
	SeoCollection *SeoCollection
}

func (sc seoController) Index(context *admin.Context) {
	context = context.NewResourceContext(sc.SeoCollection.SettingResource)
	seoInterface := sc.SeoCollection.SettingResource.NewStruct()
	context.Execute("index", seoInterface)
}
