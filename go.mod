module flamingo.me/flamingo-commerce-adapter-standalone

require (
	flamingo.me/dingo v0.1.6
	flamingo.me/flamingo-commerce/v3 v3.0.0-beta.1
	flamingo.me/flamingo/v3 v3.0.1
	github.com/RoaringBitmap/roaring v0.4.19 // indirect
	github.com/blevesearch/bleve v0.8.0
	github.com/blevesearch/go-porterstemmer v1.0.2 // indirect
	github.com/blevesearch/segment v0.0.0-20160915185041-762005e7a34f // indirect
	github.com/couchbase/vellum v0.0.0-20190823171024-95128b2d4edb // indirect
	github.com/disintegration/imaging v1.5.0
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/steveyen/gtreap v0.0.0-20150807155958-0abe01ef9be2 // indirect
	github.com/stretchr/testify v1.4.0
	golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81 // indirect
	google.golang.org/api v0.3.1
)

replace flamingo.me/flamingo/v3 => ../flamingo

replace flamingo.me/flamingo-commerce/v3 => ../flamingo-commerce

replace flamingo.me/form => ../form

replace flamingo.me/flamingo-commerce-adapter-standalone => ../flamingo-commerce-adapter-standalone
