package session

import (
	"log"

	istiov1alpha1 "github.com/aslakknutsen/istio-workspace/pkg/apis/istio/v1alpha1"
	"github.com/aslakknutsen/istio-workspace/pkg/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	versioned.Interface
	namespace string
}

// NewClient creates client to handle Session resources based on passed config
func NewClient(c versioned.Interface, namespace string) (*client, error) { //nolint[:golint] otherwise golint complains about "exported func returns unexported type *session.client, which can be annoying to use"

	return &client{namespace: namespace, Interface: c}, nil
}

var defaultClient *client

// DefaultClient creates a client based on existing kube config.
// The instance is created lazily only once and shared among all the callers
// While resolving configuration we look for .kube/config file unless KUBECONFIG env variable is set
func DefaultClient() *client { //nolint[:golint] otherwise golint complains about "exported func returns unexported type *session.client, which can be annoying to use"
	if defaultClient == nil {
		kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		var err error
		restCfg, err := kubeCfg.ClientConfig()
		if err != nil {
			log.Panicf("failed to create default client: %s", err)
		}

		c, err := versioned.NewForConfig(restCfg)
		if err != nil {
			log.Panicf("failed to create default client: %s", err)
		}

		namespace, _, err := kubeCfg.Namespace()
		if err != nil {
			log.Panicf("failed to create default client: %s", err)
		}

		defaultClient, err = NewClient(c, namespace)
		if err != nil {
			log.Panicf("failed to create default client: %s", err)
		}
	}
	return defaultClient
}

func (c *client) Create(session *istiov1alpha1.Session) error {
	if _, err := c.Interface.IstioV1alpha1().Sessions(c.namespace).Create(session); err != nil {
		return err
	}
	return nil
}

func (c *client) Delete(session *istiov1alpha1.Session) error {
	if err := c.IstioV1alpha1().Sessions(c.namespace).Delete(session.Name, &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *client) Get(sessionName string) (*istiov1alpha1.Session, error) {
	session, err := c.IstioV1alpha1().Sessions(c.namespace).Get(sessionName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return session, nil
}
