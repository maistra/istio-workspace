package dynclient

import (
	coreV1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	Namespace string
	dynClient dynamic.Interface
	clientset *kubernetes.Clientset
	mapper    meta.RESTMapper
}

func NewDefaultDynamicClient(namespace string) (*Client, error) {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return nil, err
	}

	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	return &Client{dynClient: dynClient,
			clientset: clientset,
			mapper:    rm,
			Namespace: namespace},
		nil
}

func (c *Client) Create(obj runtime.Object) error {
	err := c.createNamespaceIfNotExists()
	if err != nil {
		return err
	}

	var resourceInterface dynamic.ResourceInterface
	nsResourceInterface, err := c.resourceInterfaceFor(obj)
	if err != nil {
		return err
	}

	resourceInterface = nsResourceInterface

	switch obj.(type) {
	case *v1beta1.CustomResourceDefinition:
	case *rbacV1.ClusterRole:
	case *rbacV1.ClusterRoleBinding:
	default:
		// For all the other types we should create resources in the desired namespace
		resourceInterface = nsResourceInterface.Namespace(c.Namespace)
	}

	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}

	_, err = resourceInterface.Create(&unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})

	return err
}

func (c *Client) createNamespaceIfNotExists() error {
	nsSpec := &coreV1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: c.Namespace}}
	_, err := c.clientset.CoreV1().Namespaces().Create(nsSpec)
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *Client) resourceInterfaceFor(raw runtime.Object) (dynamic.NamespaceableResourceInterface, error) {
	gvk := raw.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := c.mapper.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, err
	}
	resource := c.dynClient.Resource(mapping.Resource)
	return resource, err
}
