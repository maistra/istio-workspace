package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/maistra/istio-workspace/version"

	"github.com/maistra/istio-workspace/pkg/openshift/parser"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/maistra/istio-workspace/pkg/openshift/dynclient"

	openshiftApi "github.com/openshift/api/template/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
)

// NewInstallCmd takes care of deploying server-side components of istio-workspace
func NewInstallCmd() *cobra.Command {
	installCmd := &cobra.Command{
		Use:          "install-operator",
		Short:        "Takes care of deploying server-side components of istio-workspace",
		SilenceUsage: true,
		RunE:         installOperator,
	}

	installCmd.Flags().StringP("namespace", "n", "istio-workspace-operator", "Target namespace to which istio-workspace operator is deployed to.")

	helpTpl := installCmd.HelpTemplate() + `
Environment variables you can override:{{range tplParams}}
{{.Name}} - {{.Description}} (default "{{.Value}}"){{end}}
`
	installCmd.SetHelpTemplate(helpTpl)
	return installCmd
}

func installOperator(cmd *cobra.Command, args []string) error { //nolint[:unparam]
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	// Propagates NAMESPACE env var which is used by templates
	if os.Getenv("NAMESPACE") == "" {
		if envErr := os.Setenv("NAMESPACE", namespace); envErr != nil {
			return envErr
		}
	}

	// Propagates IKE_IMAGE_TAG env var which is used by templates to be aligned with the version of the actual built binary
	if os.Getenv("IKE_IMAGE_TAG") == "" {
		if envErr := os.Setenv("IKE_IMAGE_TAG", version.Version); envErr != nil {
			return envErr
		}
	}

	app, err := newApplier(namespace)
	if err != nil {
		return err
	}

	if err := apply(app.applyResource, "crds/istio_v1alpha1_session_crd.yaml", "service_account.yaml", "role.yaml"); err != nil {
		return err
	}
	if err := apply(app.applyTemplate, "role_binding.yaml", "operator.yaml"); err != nil {
		return err
	}

	return nil
}

type applier struct {
	c *dynclient.Client
	d parser.DecodeFunc
}

func newApplier(namespace string) (*applier, error) { //nolint[:golint]
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
				log.Error(err, "failed creating "+strings.TrimSuffix(p, ".yaml"))
				return err
			}
			log.Info(strings.TrimSuffix(p, ".yaml") + " already exists")
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
	log.Info(fmt.Sprintf("Applying '%s' in namespace '%s' [Name: %s; Kind: %s]",
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
		log.Info(fmt.Sprintf("Applying '%s' in namespace '%s' [Name: %s; Kind: %s]", strings.TrimSuffix(templatePath, ".yaml"),
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
