package session

import (
	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	client    *versioned.Clientset
	config    clientcmd.ClientConfig
	namespace string
}

func NewDefaultClient() (*client, error) { //nolint[:golint]

	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	namespace, _, err := kubeCfg.Namespace()
	if err != nil {
		return nil, err
	}

	clientCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return nil, err
	}

	c, err := versioned.NewForConfig(clientCfg)

	if err != nil {
		return nil, err
	}

	return &client{namespace: namespace, client: c, config: kubeCfg}, nil
}

func (c *client) Create(session *istiov1alpha1.Session) error {

	if _, err := c.client.IstioV1alpha1().Sessions(c.namespace).Create(session); err != nil {
		return err
	}

	return nil
}

func (c *client) Delete(session *istiov1alpha1.Session) error {
	if err := c.client.IstioV1alpha1().Sessions(c.namespace).Delete(session.Name, &metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}

func (c *client) getSession(sessionName string) (*istiov1alpha1.Session, error) {
	session, err := c.client.IstioV1alpha1().Sessions(c.namespace).Get(sessionName, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return session, nil
}
