module github.com/maistra/istio-workspace

go 1.14

require (
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/go-cmd/cmd v1.3.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/golang/protobuf v1.4.3
	github.com/google/go-cmp v0.5.3 // indirect
	github.com/google/go-github/v32 v32.1.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hashicorp/go-multierror v1.1.0
	github.com/joho/godotenv v1.3.0
	github.com/kisielk/errcheck v1.6.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mikefarah/yq/v4 v4.6.1
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/openshift/api v0.0.0-20200527184302-a843dc3262a0
	github.com/operator-framework/operator-lib v0.3.0
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/prometheus/client_golang v1.9.0
	github.com/sabhiram/go-gitignore v0.0.0-20180611051255-d3107576ba94 // v1.0.2
	github.com/spf13/afero v1.5.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	go.uber.org/goleak v1.1.10
	go.uber.org/zap v1.16.0
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5 // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/tools v0.1.0
	google.golang.org/grpc v1.36.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/h2non/gock.v1 v1.0.16
	gopkg.in/ini.v1 v1.62.0 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	istio.io/api v0.0.0-20210218044411-561dc276d04d
	istio.io/client-go v1.9.1
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
	k8s.io/code-generator v0.20.4
	sigs.k8s.io/controller-runtime v0.7.0
	sigs.k8s.io/controller-tools v0.5.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	gopkg.in/fsnotify.v1 v1.4.9 => github.com/fsnotify/fsnotify v1.4.9
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.7.0
)
