# Qor SEO Meta Module

SEO Meta module provide a administrator interface for End User to change html page's title, and description tag content.
In title and description you can embed variables that provided by Developer.


## Developer side usage

Developer will define a list of page SEO meta settings that he want the End User to use. He may provide one settings for a group of pages that has the same type. For example: All Product Detail pages, That End User only need to provide one single template configuration by utilizing Product Name, Product Code variables that developer provided.

### Example

```go
type Seo struct {
	gorm.Model
	SiteName    string
	SiteHost    string
	HomePage    seo.Setting `seo:"Topic,EmailTitle,EmailContent,EmailAddress"`
	ProductPage seo.Setting `seo:"Name"`
}
```
Developer will define the above Go struct, It means End User can setup SiteName, SiteHost as literially string. and Could setup HomePage with variables `Topic`, `EmailTitle`, `EmailContent`, `EmailAddress`. and so on.


## End User side usage

The above definiation for Developer will generate the following Qor Adminstration backend. That End user can change text and variable of the whole site title, and description:

![Administration UI](https://raw.githubusercontent.com/qor/seo/master/images/qor_meta.png)
