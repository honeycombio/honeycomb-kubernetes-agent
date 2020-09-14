module github.com/honeycombio/honeycomb-kubernetes-agent

go 1.12

require (
	github.com/boltdb/bolt v1.3.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/honeycombio/dynsampler-go v0.2.1
	github.com/honeycombio/gonx v1.3.1-0.20171118020637-f9b2468e9ef8
	github.com/honeycombio/honeytail v1.1.4
	github.com/honeycombio/libhoney-go v1.13.0
	github.com/honeycombio/urlshaper v0.0.0-20170302202025-2baba9ae5b5f
	github.com/hpcloud/tail v1.0.1-0.20170814160653-37f427138745
	github.com/jessevdk/go-flags v1.4.0
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515
	github.com/mitchellh/mapstructure v1.3.3
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.19.1
	k8s.io/client-go v0.18.8
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
)
