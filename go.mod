module github.com/maistra/istio-workspace

go 1.15

require (
	k8s.io/code-generator kubernetes-1.16.0
	k8s.io/api kubernetes-1.16.0
	k8s.io/apiextensions-apiserver kubernetes-1.16.0
	k8s.io/apimachinery kubernetes-1.16.0
	k8s.io/client-go kubernetes-1.16.0
    sigs.k8s.io/controller-runtime v0.5.0

    istio.io/api release-1.4
    istio.io/client-go release-1.4

    github.com/operator-framework/operator-sdk v0.17.2
    github.com/openshift/api release-4.3
    github.com/go-cmd/cmd v1.1.0
    github.com/hashicorp/go-multierror v1.1.0

    github.com/onsi/gomega v1.10.3

    github.com/onsi/ginkgo v1.14.2

    github.com/joho/godotenv v1.3.0

    go.uber.org/goleak v1.1.10

    github.com/sabhiram/go-gitignore v0.0.0-20180611051255-d3107576ba94 // v1.0.2
	github.com/spf13/viper v1.7.1

    github.com/spf13/cobra v1.1.1

    github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510

    gopkg.in/h2non/gock.v1 v1.0.15

    github.com/google/go-github/v32 v32.1.0

    k8s.io/kube-openapi release-1.16

    github.com/coreos/prometheus-operator v0.34.0



    //## Locking k8s/{klog,utils,gengo} to before klog@v2.0.0 as otherwise we are facing https://github.com/kubernetes/klog/issues/138
    //[[override]]
    //  name = "k8s.io/utils"
    //  revision = "5770800c2500f42361fa90f2d5df947d2c5db138"
	//
    //[[override]]
    //  name = "k8s.io/klog"
    //  version = "v1.0.0"
	//
    //[[override]]
    //  name = "k8s.io/gengo"
    //  revision = "793b05dca9b871fdc15aeaff1f201e141ef5afa7"
)

replace gopkg.in/fsnotify.v1 v1.4.9 => github.com/fsnotify/fsnotify v1.4.9
