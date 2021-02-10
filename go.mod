module github.com/lf-edge/eden

go 1.15

require (
	cloud.google.com/go v0.76.0 // indirect
	github.com/Azure/azure-sdk-for-go v42.3.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.10.2 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/amitbet/vncproxy v0.0.0-20200118084310-ea8f9b510913
	github.com/aws/aws-sdk-go v1.31.6 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bshuster-repo/logrus-logstash-hook v1.0.0 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/containerd/cgroups v0.0.0-20210114181951-8a68de567b68 // indirect
	github.com/containerd/containerd v1.4.3
	github.com/containerd/continuity v0.0.0-20210208174643-50096c924a4e // indirect
	github.com/containerd/fifo v0.0.0-20210129194248-f8e8fdba47ef // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.0.0-20210210143742-bdc018f2992a // indirect
	github.com/deislabs/oras v0.10.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v20.10.3+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.10.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/garyburd/redigo v1.6.2 // indirect
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/go-redis/redis/v7 v7.4.0
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.4.3
	github.com/google/go-containerregistry v0.4.0
	github.com/google/uuid v1.2.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/lf-edge/adam v0.0.0-20210203004636-0b72a46462c4
	github.com/lf-edge/eden/eserver v0.0.0-20210209103608-121a7ab1db51
	github.com/lf-edge/edge-containers v0.0.0-20210210095907-df5959d36b82
	github.com/lf-edge/eve/api/go v0.0.0-20210209052624-0ef611252120
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40 // indirect
	github.com/mcuadros/go-lookup v0.0.0-20200831155250-80f87a4fa5ee
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635
	github.com/nerd2/gexto v0.0.0-20190529073929-39468ec063f6
	github.com/opencontainers/selinux v1.8.0 // indirect
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/prometheus/common v0.15.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/rogpeppe/go-internal v1.6.0
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.2
	github.com/spf13/viper v1.7.1
	github.com/vdemeester/k8s-pkg-credentialprovider v1.18.1-0.20201019120933-f1d16962a4db // indirect
	github.com/vmihailenco/msgpack/v4 v4.3.12 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	go.opencensus.io v0.22.6 // indirect
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	golang.org/x/oauth2 v0.0.0-20210201163806-010130855d6c
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	google.golang.org/api v0.39.0
	google.golang.org/genproto v0.0.0-20210207032614-bba0dbe2a9ea // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/errgo.v2 v2.1.0
	gopkg.in/yaml.v2 v2.4.0
	gotest.tools/gotestsum v1.6.1 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
)

replace github.com/lf-edge/eden/eserver => ./eserver

replace github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

replace github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
