# SEO

This SEO library allows for the management of and injection of dynamic data into HTML tags for the purpose of Search Engine Optimisation. Using the [QOR Admin](https://github.com/qor/admin) interface, an administrator can easily manage the content of an HTML page's title, description, and meta tags.

[![GoDoc](https://godoc.org/github.com/qor/seo?status.svg)](https://godoc.org/github.com/qor/seo)

## Definition

Define SEO settings struct with `Site-wide Settings` and `Page Related Settings`, for example:

```go
type SEO struct {
	gorm.Model
	SiteName     string  // Site-wide Settings
	SiteHost     string  // Site-wide Settings
  HomePage     seo.Setting
  ProductPage  seo.Setting `seo:"ProductName,ProductCode"` // Page Related Variables [ProductName, ProductCode]
  CategoryPage seo.Setting `seo:"CategoryName"` // Page Related Variables [CategoryName]
}

// The `SEO` struct is a normal GORM-backend model, need to run migration before using it
db.AutoMigrate(&SEO{})

// Add `SEO` to QOR Admin Interface
Admin.AddResource(&SEO{})
```

[Online SEO Setting Demo For Qor Example](http://demo.getqor.com/admin/seo_setting)

## Usage

```go
var SEOSetting SEO
db.First(&SEOSetting)

// render home page's meta tags
SEOSetting.HomePage.Render(SEOSetting)

// render product page's meta tags
var product Product
db.First(&product, "code = ?", "L1212")
SEOSetting.ProductPage.Render(SEOSetting, product)
```

## Structured Data

```go
// micro search
seo.MicroSearch{
  URL:    "http://demo.getqor.com",
  Target: "http://demo.getqor.com/search?q=",
}.Render()

// micro contact
seo.MicroContact{
  URL:         "http://demo.getqor.com",
  Telephone:   "080-0012-3232",
  ContactType: "Customer Service",
}.Render()

// micro product
seo.MicroProduct{
  Name: "Kenmore White 17 Microwave",
  Image: "http://getqor.com/source/images/qor-logo.png",
  Description: "0.7 cubic feet countertop microwave. Has six preset cooking categories and convenience features like Add-A-Minute and Child Lock."
  BrandName: "ThePlant",
  SKU: "L1212",
  PriceCurrency: "USD",
  Price: 100,
  SellerName: "ThePlant",
}.Render()
```

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
