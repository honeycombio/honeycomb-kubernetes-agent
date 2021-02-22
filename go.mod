module github.com/honeycombio/honeycomb-kubernetes-agent

go 1.12

require (
	github.com/boltdb/bolt v1.3.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/honeycombio/dynsampler-go v0.2.1
	github.com/honeycombio/gonx v1.3.1-0.20171118020637-f9b2468e9ef8
	github.com/honeycombio/honeytail v1.2.0
	github.com/honeycombio/libhoney-go v1.15.2
	github.com/honeycombio/urlshaper v0.0.0-20170302202025-2baba9ae5b5f
	github.com/hpcloud/tail v1.0.1-0.20170814160653-37f427138745
	github.com/jessevdk/go-flags v1.4.0
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3 // indirect
	github.com/stretchr/testify v1.7.0
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.2
	k8s.io/kubernetes v1.12.0
)
