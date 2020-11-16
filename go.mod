module github.com/maistra/istio-workspace

go 1.15

require (
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-cmd/cmd v1.1.0
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1
	github.com/golang/protobuf v1.4.2
	github.com/google/go-github/v32 v32.1.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hashicorp/go-multierror v1.1.0
	github.com/joho/godotenv v1.3.0
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/openshift/api v0.0.0-20200527184302-a843dc3262a0
	github.com/operator-framework/operator-sdk v0.17.2
	github.com/prometheus/client_golang v1.5.1
	github.com/sabhiram/go-gitignore v0.0.0-20180611051255-d3107576ba94 // v1.0.2
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	go.uber.org/goleak v1.1.10
	go.uber.org/zap v1.15.0
	golang.org/x/net v0.0.0-20201006153459-a7d1128ccaa0
	google.golang.org/grpc v1.27.0
	google.golang.org/protobuf v1.23.0
	gopkg.in/h2non/gock.v1 v1.0.15
	gopkg.in/yaml.v2 v2.3.0
	istio.io/api v0.0.0-20200107183329-ed4b507c54e1
	istio.io/client-go v0.0.0-20200107185429-9053b0f86b03
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.17.4
	k8s.io/kube-openapi v0.0.0-20200410163147-594e756bea31 // indirect
	sigs.k8s.io/controller-runtime v0.5.2
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.34.0
	gopkg.in/fsnotify.v1 v1.4.9 => github.com/fsnotify/fsnotify v1.4.9
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.0
)
