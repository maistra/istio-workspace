package install

import (
	"fmt"
	"os"
	"strings"

	"github.com/maistra/istio-workspace/pkg/assets"

	"github.com/maistra/istio-workspace/pkg/client/dynclient"
	"github.com/maistra/istio-workspace/pkg/log"
	"github.com/maistra/istio-workspace/pkg/openshift/parser"
	"github.com/maistra/istio-workspace/version"

	"github.com/go-logr/logr"
	openshiftApi "github.com/openshift/api/template/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
)

var logger = func() logr.Logger {
	return log.Log.WithValues("type", "install")
}

// NewCmd takes care of deploying server-side components of istio-workspace.
func NewCmd() *cobra.Command {
	installCmd := &cobra.Command{
		Use:          "install",
		Short:        "Takes care of deploying server-side components of istio-workspace",
		SilenceUsage: true,
		RunE:         installOperator,
	}

	installCmd.Flags().StringP("namespace", "n", "istio-workspace-operator", "Target namespace to which istio-workspace operator is deployed to.")
	installCmd.Flags().BoolP("local", "l", false, "Install as local to the namespace only")

	helpTpl := installCmd.HelpTemplate() + `
Environment variables you can override:{{range tplParams}}
{{.Name}} - {{.Description}} (default "{{.Value}}"){{end}}
`
	installCmd.SetHelpTemplate(helpTpl)
	return installCmd
}

func installOperator(cmd *cobra.Command, args []string) error { //nolint:gocyclo //reason cyclo can be skipped for the sake of readability
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	local, err := cmd.Flags().GetBool("local")
	if err != nil {
		return err
	}

	if local && !cmd.Flags().Changed("namespace") {
		kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		if namespace, _, err = kubeCfg.Namespace(); err != nil {
			logger().Error(err, fmt.Sprintf("failed to read k8s config file [%s]. "+
				"use --namespace to define existing target namespace", kubeCfg.ConfigAccess().GetExplicitFile()))
			return err
		}
	}

	// Propagates NAMESPACE env var which is used by templates
	if _, found := os.LookupEnv("NAMESPACE"); !found {
		if envErr := os.Setenv("NAMESPACE", namespace); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_VERSION env var which is used by templates to be aligned with the version of the actual built binary
	if _, found := os.LookupEnv("IKE_VERSION"); !found {
		if envErr := os.Setenv("IKE_VERSION", version.Version); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_IMAGE_TAG env var which is used by templates to be aligned with the version of the actual built binary
	if _, found := os.LookupEnv("IKE_IMAGE_TAG"); !found {
		if envErr := os.Setenv("IKE_IMAGE_TAG", version.Version); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_DOCKER_REGISTRY env var which is used by templates to build the image url
	if _, found := os.LookupEnv("IKE_DOCKER_REGISTRY"); !found {
		if envErr := os.Setenv("IKE_DOCKER_REGISTRY", "quay.io"); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_DOCKER_REPOSITORY env var which is used by templates to build the image url
	if _, found := os.LookupEnv("IKE_DOCKER_REPOSITORY"); !found {
		if envErr := os.Setenv("IKE_DOCKER_REPOSITORY", "maistra"); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_IMAGE_NAME env var which is used by templates to build the image url
	if _, found := os.LookupEnv("IKE_IMAGE_NAME"); !found {
		if envErr := os.Setenv("IKE_IMAGE_NAME", "istio-workspace"); envErr != nil {
			return envErr
		}
	}

	app, err := newApplier(namespace)
	if err != nil {
		return err
	}
	resources := []string{"crds/maistra.io_sessions.yaml", "service_account.yaml", "cluster_role.yaml"}
	templates := []string{"cluster_role_binding.yaml", "operator.tpl.yaml"}

	if local {
		resources = []string{"crds/maistra.io_sessions.yaml", "service_account.yaml", "role.yaml"}
		templates = []string{"role_binding.yaml", "operator.tpl.yaml"}

		if envErr := os.Setenv("WATCH_NAMESPACE", namespace); envErr != nil {
			return envErr
		}
	}

	if err := apply(app.applyResource, resources...); err != nil {
		return err
	}
	if err := apply(app.applyTemplate, templates...); err != nil {
		return err
	}

	logger().Info("Installing operator", "namespace", os.Getenv("NAMESPACE"),
		"image", fmt.Sprintf("%s/%s/%s:%s", os.Getenv("IKE_DOCKER_REGISTRY"),
			os.Getenv("IKE_DOCKER_REPOSITORY"),
			os.Getenv("IKE_IMAGE_NAME"),
			os.Getenv("IKE_IMAGE_TAG")),
		"version", version.Version)

	return nil
}

type applier struct {
	c *dynclient.Client
	d parser.DecodeFunc
}

func newApplier(namespace string) (*applier, error) {
	client, err := dynclient.NewDefaultDynamicClient(namespace)
	if err != nil {
		return nil, err
	}
	decode, err := parser.Decoder()
	if err != nil {
		return nil, err
	}

	return &applier{c: client, d: decode}, nil
}

func apply(a func(path string) error, paths ...string) error {
	for _, p := range paths {
		if err := a(p); err != nil {
			if !errors.IsAlreadyExists(err) {
				logger().Error(err, "failed creating "+strings.TrimSuffix(p, ".yaml"))
				return err
			}
			logger().Info(strings.TrimSuffix(p, ".yaml") + " already exists")
		}
	}
	return nil
}

func (app *applier) applyResource(resourcePath string) error {
	rawCrd, err := assets.Load("deploy/" + resourcePath)
	if err != nil {
		return err
	}
	crd, err := parser.Parse(rawCrd)
	if err != nil {
		return err
	}
	kind := crd.GetObjectKind()
	logger().Info(fmt.Sprintf("Applying '%s' in namespace '%s' [Name: %s; Kind: %s]",
		strings.TrimSuffix(resourcePath, ".yaml"),
		app.c.Namespace,
		name(crd, resourcePath),
		kind.GroupVersionKind().Kind))
	err = app.c.Create(crd)
	return err
}

func name(object runtime.Object, fallback string) (name string) {
	accessor := meta.NewAccessor()
	name, err := accessor.Name(object)
	if err != nil {
		name = fallback
	}
	return
}

func (app *applier) applyTemplate(templatePath string) error {
	yaml, err := parser.ProcessTemplateUsingEnvVars("deploy/" + templatePath)
	if err != nil {
		return err
	}

	tpl, err := parser.Parse(yaml)
	if err != nil {
		return err
	}

	r := tpl.(*openshiftApi.Template)
	for _, obj := range r.Objects {
		object, gav, err := app.d(obj.Raw, nil, nil)
		if err != nil {
			return err
		}
		logger().Info(fmt.Sprintf("Applying '%s' in namespace '%s' [Name: %s; Kind: %s]", strings.TrimSuffix(templatePath, ".yaml"),
			app.c.Namespace,
			name(object, templatePath),
			gav.Kind))
		err = app.c.Create(object)
		if err != nil {
			return err
		}
	}
	return nil
}
