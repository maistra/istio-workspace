package dynclient

import (
	"context"

	"emperror.dev/errors"
	coreV1 "k8s.io/api/core/v1"
	rbacV1 "k8s.io/api/rbac/v1"
	extV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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
	clientset kubernetes.Interface
	mapper    meta.RESTMapper
}

func NewClient(dynClient dynamic.Interface, clientset kubernetes.Interface, mapper meta.RESTMapper) Client {
	return Client{
		Namespace: "",
		dynClient: dynClient,
		clientset: clientset,
		mapper:    mapper,
	}
}

// NewDefaultDynamicClient creates dynamic client for given ns.
func NewDefaultDynamicClient(namespace string, createNs bool) (*Client, error) {
	kubeCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restCfg, err := kubeCfg.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed configuring k8s client")
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating k8s clientset")
	}

	dynClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed creating k8s dynamic client")
	}

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		return nil, errors.Wrap(err, "failed obtaining APIGroupResources")
	}

	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	client := Client{dynClient: dynClient,
		clientset: clientset,
		mapper:    rm,
		Namespace: namespace}

	if createNs {
		err = client.createNamespaceIfNotExists()
	}

	return &client, err
}

func (c *Client) Dynamic() dynamic.Interface {
	return c.dynClient
}

func (c *Client) Delete(obj runtime.Object) error {
	resourceInterface, err := c.resourceInterfaceFor(obj)
	if err != nil {
		return errors.Wrap(err, "failed creating resource interface")
	}

	name, err := meta.NewAccessor().Name(obj)
	if err != nil {
		return errors.Wrap(err, "failed obtaining name")
	}

	err = resourceInterface.Delete(context.Background(), name, metav1.DeleteOptions{})

	return errors.Wrap(err, "failed deleting object")
}

func (c *Client) Create(obj runtime.Object) error {
	err := c.createNamespaceIfNotExists()
	if err != nil {
		return err
	}

	resourceInterface, err := c.resourceInterfaceFor(obj)
	if err != nil {
		return errors.Wrap(err, "failed creating resource interface")
	}

	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return errors.Wrap(err, "failed converting object to unstructured object")
	}

	_, err = resourceInterface.Create(context.Background(), &unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})

	return errors.Wrap(err, "failed creating object")
}

func (c *Client) resourceInterfaceFor(obj runtime.Object) (dynamic.ResourceInterface, error) {
	var resourceInterface dynamic.ResourceInterface
	nsResourceInterface, err := c.resourceNsInterfaceFor(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed obtaining resource interface")
	}

	resourceInterface = nsResourceInterface

	switch obj.(type) {
	case *extV1.CustomResourceDefinition:
	case *rbacV1.ClusterRole:
	case *rbacV1.ClusterRoleBinding:
	default:
		// For all the other types we should create resources in the desired namespace
		resourceInterface = nsResourceInterface.Namespace(c.Namespace)
	}

	return resourceInterface, nil
}

func (c *Client) createNamespaceIfNotExists() error {
	nsSpec := &coreV1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: c.Namespace}}
	_, err := c.clientset.CoreV1().Namespaces().Create(context.Background(), nsSpec, metav1.CreateOptions{})
	if k8sErrors.IsAlreadyExists(err) {
		return nil
	}

	return errors.Wrap(err, "failed creating new namespace")
}

func (c *Client) resourceNsInterfaceFor(raw runtime.Object) (dynamic.NamespaceableResourceInterface, error) {
	gvk := raw.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := c.mapper.RESTMapping(gk, gvk.Version)
	if err != nil {
		return nil, errors.Wrap(err, "failed mapping runtime object")
	}
	resource := c.dynClient.Resource(mapping.Resource)

	return resource, nil
}
