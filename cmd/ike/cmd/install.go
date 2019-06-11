package cmd

import (
	"github.com/maistra/istio-workspace/pkg/openshift"

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

	return installCmd
}

func installOperator(cmd *cobra.Command, args []string) error { //nolint[:unparam]
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	app, err := newApplier(namespace)
	if err != nil {
		return err
	}
	if err := app.applyResource("crds/istio_v1alpha1_session_crd.yaml"); err != nil {
		return err
	}
	if err := app.applyResource("service_account.yaml"); err != nil {
		return err
	}
	if err := app.applyResource("role.yaml"); err != nil {
		return err
	}
	if err := app.applyTemplate("role_binding.yaml"); err != nil {
		return err
	}
	if err := app.applyTemplate("operator.yaml"); err != nil {
		return err
	}

	return nil
}

type applier struct {
	c *openshift.Client
	d openshift.DecodeFunc
}

func (app *applier) applyResource(resourcePath string) error {
	rawCrd, err := openshift.Load("deploy/istio-workspace/" + resourcePath)
	if err != nil {
		return err
	}
	crd, err := openshift.Parse(rawCrd)
	if err != nil {
		return err
	}
	err = app.c.Create(crd)
	return err
}

func (app *applier) applyTemplate(templatePath string) error {
	yaml, err := openshift.ProcessTemplateUsingEnvVars("deploy/istio-workspace/" + templatePath)
	if err != nil {
		return err
	}

	rawRoleBinding, err := openshift.Parse(yaml)
	if err != nil {
		return err
	}

	r := rawRoleBinding.(*openshiftApi.Template)
	for _, obj := range r.Objects {
		object, _, err := app.d(obj.Raw, nil, nil)
		if err != nil {
			return err
		}
		err = app.c.Create(object)
		if err != nil {
			return err
		}
	}
	return nil
}

func newApplier(namespace string) (*applier, error) { //nolint[:golint]
	client, err := openshift.NewDefaultDynamicClient(namespace)
	if err != nil {
		return nil, err
	}
	decode, err := openshift.Decoder()
	if err != nil {
		return nil, err
	}

	return &applier{c: client, d: decode}, nil
}
