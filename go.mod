module flamingo.me/flamingo-commerce-adapter-standalone

require (
	flamingo.me/dingo v0.2.9
	flamingo.me/flamingo-commerce/v3 v3.0.0-beta.1
	flamingo.me/flamingo/v3 v3.2.1-0.20200812074650-142034e9fe96
	github.com/Masterminds/sprig v2.16.0+incompatible
	github.com/RoaringBitmap/roaring v0.4.19 // indirect
	github.com/blevesearch/bleve v0.8.0
	github.com/blevesearch/go-porterstemmer v1.0.2 // indirect
	github.com/blevesearch/segment v0.0.0-20160915185041-762005e7a34f // indirect
	github.com/couchbase/vellum v0.0.0-20190823171024-95128b2d4edb // indirect
	github.com/disintegration/imaging v1.5.0
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/fzipp/gocyclo v0.0.0-20150627053110-6acd4345c835 // indirect
	github.com/gordonklaus/ineffassign v0.0.0-20200809085317-e36bfde3bb78 // indirect
	github.com/jordan-wright/email v0.0.0-20200602115436-fd8a7622303e
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/steveyen/gtreap v0.0.0-20150807155958-0abe01ef9be2 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/vanng822/go-premailer v0.0.0-20191214114701-be27abe028fe
	golang.org/x/image v0.0.0-20180708004352-c73c2afc3b81 // indirect
	google.golang.org/api v0.19.0
)

replace flamingo.me/flamingo/v3 => ../flamingo

replace flamingo.me/flamingo-commerce/v3 => ../flamingo-commerce

replace flamingo.me/form => ../form

replace flamingo.me/flamingo-commerce-adapter-standalone => ../flamingo-commerce-adapter-standalone

go 1.13
