# csvcommerce

This module provides:
* a IndexUpdater for "commercesearch" module
* an image controller that shows product images


## Usage for products:

Place a `products.csv`  and the product images in a folder - e.g. "ressources".

The images are served with the imagecontroller under the default route `/image/:size/:filename`
The ":size" parameter is in the form "widthxheight" - for example "200x" would scale the image to a width of 200px and the height is adjusted automatically.
The allowed imagesizes need to be configured (`flamingo-commerce-adapter-standalone.csvCommerce.allowedImageResizeParamaters`)

See the commerce-demo-carotene Demoshop for an example usage.

Categories are only indexed if a category.csv is given.

## Configuration

```
flamingoCommerceAdapterStandalone: 
	csvindexing: 
		productCsvPath: "ressources/products/products.csv"
		categoryCsvPath: "ressources/products/category.csv"
		locale: "en_GB"
		currency: "€"
		allowedImageResizeParamaters: "200x,300x,400x,x200,x300"
	}
```

 
## CSV Format

Use UTF8 encoded CSV with optional quotes and comma (,) as seperator. 


### Product CSV

The text fields that are normaly localized need to be postfixed with the configured local.
So if you have the configuration `flamingoCommerceAdapterStandalone.csvindexing.locale: en_GB`
a product title fieldname in the CSV need to be `title-en_GB`.

Price fields need to be postfixed with currency name.
Asset and Images paths need to be relative to the CSV folder.

Must have fields:
* marketplaceCode (the unique idendifier of the product)
* title-LOCALEPREFIX (the title)
* productType ("simple" or "configurable")
* price-CURRENCY 

Optional:
* specialPrice-CURRENCY
* categories (comma separated references to categories. Using the category code as idendifier)
* retailerCode (reference to the retailer / vendor of the product)
* shortDescription-LOCALEPREFIX
* description-LOCALEPREFIX
* metaKeywords-LOCALEPREFIX (comma separated keywords)
* listImage
* thumbnailImage
* detailImage01,detailImage02 ... detailImage10
* ??? - any other field will be added as a product attribute

For configurable product types:
* variantVariationAttributes (The attribute)
* CONFIGURABLE-products (comma separated references to other products (use marketplacecode as id))

```
"sku","gtin","name","title-en_GB","description-en_GB","shortDescription-en_GB","listTeaser-en_GB","price-GBP","specialPrice-GBP","specialOffer","metaTitle-en_GB","metaKeywords-en_GB","detailImage01","listImage","thumbnailImage","manufacturerColor-en_GB","categories","retailerCode","family","marketplaceCode","brandCode","productType","clothing_size","colour","CONFIGURABLE-products","variantVariationAttributes"
"hellokitty-s-red",98,"Hello Kitty S Red","Hello Kitty S Red","Hello Kitty S Red description","Hello Kitty is great",,"30.00",,,,,"productfiles/images/sanni-sahil-1173038-unsplash.jpg","productfiles/images/sanni-sahil-1173038-unsplash.jpg","productfiles/images/sanni-sahil-1173038-unsplash.jpg",,"clothing","aoepeople","cross_segment","hellokitty-s-red","kitty","simple","S","Red",,
```


### Category CSV
"code","parent","label-en_GB"
"master","master","master"
"clothing","master","Clothing"
"accessories","master","accessories"