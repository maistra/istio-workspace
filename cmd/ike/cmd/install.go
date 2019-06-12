package cmd

import (
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
	d dynclient.DecodeFunc
}

func newApplier(namespace string) (*applier, error) { //nolint[:golint]
	client, err := dynclient.NewDefaultDynamicClient(namespace)
	if err != nil {
		return nil, err
	}
	decode, err := dynclient.Decoder()
	if err != nil {
		return nil, err
	}

	return &applier{c: client, d: decode}, nil
}

func apply(a func(path string) error, paths ...string) error {
	for _, p := range paths {
		log.Info("applying " + p)
		if err := a(p); err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Error(err, "failed creating "+p)
				return err
			}
			log.Info(p + "already exists")
		}
	}
	return nil
}

func (app *applier) applyResource(resourcePath string) error {
	rawCrd, err := dynclient.Load("deploy/istio-workspace/" + resourcePath)
	if err != nil {
		return err
	}
	crd, err := dynclient.Parse(rawCrd)
	if err != nil {
		return err
	}
	err = app.c.Create(crd)
	return err
}

func (app *applier) applyTemplate(templatePath string) error {
	yaml, err := dynclient.ProcessTemplateUsingEnvVars("deploy/istio-workspace/" + templatePath)
	if err != nil {
		return err
	}

	rawRoleBinding, err := dynclient.Parse(yaml)
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
