package install

import (
	"fmt"
	"os"
	"strings"

	"github.com/maistra/istio-workspace/pkg/log"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/maistra/istio-workspace/version"

	"github.com/maistra/istio-workspace/pkg/openshift/parser"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/maistra/istio-workspace/pkg/client/dynclient"

	openshiftApi "github.com/openshift/api/template/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
)

var logger = log.CreateOperatorAwareLogger("cmd").WithValues("type", "install")

// NewCmd takes care of deploying server-side components of istio-workspace.
func NewCmd() *cobra.Command {
	installCmd := &cobra.Command{
		Use:          "install-operator",
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
			logger.Error(err, fmt.Sprintf("failed to read k8s config file [%s]. "+
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

	// Propagates IKE_IMAGE_TAG env var which is used by templates to be aligned with the version of the actual built binary
	if _, found := os.LookupEnv("IKE_IMAGE_TAG"); !found {
		if envErr := os.Setenv("IKE_IMAGE_TAG", version.Version); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_VERSION env var which is used by templates to be aligned with the version of the actual built binary
	if _, found := os.LookupEnv("IKE_VERSION"); !found {
		if envErr := os.Setenv("IKE_VERSION", version.Version); envErr != nil {
			return envErr
		}
	}

	app, err := newApplier(namespace)
	if err != nil {
		return err
	}
	resources := []string{"crds/istio_v1alpha1_session_crd.yaml", "service_account.yaml", "role.yaml"}
	templates := []string{"role_binding.yaml", "operator.yaml"}

	if local {
		resources = []string{"crds/istio_v1alpha1_session_crd.yaml", "service_account.yaml", "role_local.yaml"}
		templates = []string{"role_binding_local.yaml", "operator.yaml"}

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

	logger.Info("Installing operator", "namespace", os.Getenv("NAMESPACE"),
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
				logger.Error(err, "failed creating "+strings.TrimSuffix(p, ".yaml"))
				return err
			}
			logger.Info(strings.TrimSuffix(p, ".yaml") + " already exists")
		}
	}
	return nil
}

func (app *applier) applyResource(resourcePath string) error {
	rawCrd, err := parser.Load("deploy/istio-workspace/" + resourcePath)
	if err != nil {
		return err
	}
	crd, err := parser.Parse(rawCrd)
	if err != nil {
		return err
	}
	kind := crd.GetObjectKind()
	logger.Info(fmt.Sprintf("Applying '%s' in namespace '%s' [Name: %s; Kind: %s]",
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
	yaml, err := parser.ProcessTemplateUsingEnvVars("deploy/istio-workspace/" + templatePath)
	if err != nil {
		return err
	}

	rawRoleBinding, err := parser.Parse(yaml)
	if err != nil {
		return err
	}

	r := rawRoleBinding.(*openshiftApi.Template)
	for _, obj := range r.Objects {
		object, gav, err := app.d(obj.Raw, nil, nil)
		if err != nil {
			return err
		}
		logger.Info(fmt.Sprintf("Applying '%s' in namespace '%s' [Name: %s; Kind: %s]", strings.TrimSuffix(templatePath, ".yaml"),
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
