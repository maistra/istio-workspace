package session

import (
	"context"

	istiov1alpha1 "github.com/maistra/istio-workspace/api/maistra/v1alpha1"
	"github.com/maistra/istio-workspace/pkg/client/clientset/versioned"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// Client interacts with the k8s api server.
type Client struct {
	versioned.Interface
	namespace string
}

// NewClient creates client to handle Session resources based on passed config.
func NewClient(c versioned.Interface, namespace string) (*Client, error) {
	return &Client{namespace: namespace, Interface: c}, nil
}

var defaultClient *Client

// DefaultClient creates a client based on existing kube config.
// The instance is created lazily only once and shared among all the callers
// While resolving configuration we look for .kube/config file unless KUBECONFIG env variable is set
// If namespace parameter is empty default one from the current context is used.
func DefaultClient(namespace string) (*Client, error) {
	if defaultClient == nil {
		c2, err2 := createDefaultClient(namespace)
		if err2 != nil {
			return c2, err2
		}
	}
	return defaultClient, nil
}

func createDefaultClient(namespace string) (*Client, error) {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	var err error
	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		logger().Error(err, "failed to create default client")
		return nil, err
	}

	c, err := versioned.NewForConfig(restCfg)
	if err != nil {
		logger().Error(err, "failed to create default client")
		return nil, err
	}

	if namespace == "" {
		namespace, _, err = kubeCfg.Namespace()
		if err != nil {
			logger().Error(err, "failed to create default client")
			return nil, err
		}
	}
	defaultClient, err = NewClient(c, namespace)
	if err != nil {
		logger().Error(err, "failed to create default client")
		return nil, err
	}
	return defaultClient, nil
}

// Create creates a session instance in a cluster.
func (c *Client) Create(session *istiov1alpha1.Session) error {
	if _, err := c.Interface.MaistraV1alpha1().Sessions(c.namespace).Create(context.Background(), session, metav1.CreateOptions{}); err != nil {
		return err
	}
	return nil
}

// Update updates a session instance in a cluster.
func (c *Client) Update(session *istiov1alpha1.Session) error {
	if _, err := c.Interface.MaistraV1alpha1().Sessions(c.namespace).Update(context.Background(), session, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

// Delete deletes a session instance in a cluster.
func (c *Client) Delete(session *istiov1alpha1.Session) error {
	return c.MaistraV1alpha1().Sessions(c.namespace).Delete(context.Background(), session.Name, metav1.DeleteOptions{})
}

// Get retrieves details of the Session instance matching passed name.
func (c *Client) Get(sessionName string) (*istiov1alpha1.Session, error) {
	session, err := c.MaistraV1alpha1().Sessions(c.namespace).Get(context.Background(), sessionName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return session, nil
}
