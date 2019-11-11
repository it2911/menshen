package ext

import (
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type GroupExtInterface interface {
	List(opts metav1.ListOptions) (*menshenv1beta1.GroupExtList, error)
	Get(resourceName string, namespace string) (*menshenv1beta1.GroupExt, error)
}

type groupextClient struct {
	restClient rest.Interface
}

func (c *groupextClient) Get(resourceName string, namespace string) (*menshenv1beta1.GroupExt, error) {
	result := menshenv1beta1.GroupExt{}
	err := c.restClient.
		Get().
		Resource("groupexts").
		Name(resourceName).
		Namespace(namespace).
		Do().
		Into(&result)

	return &result, err
}

func (c *groupextClient) List(opts metav1.ListOptions) (*menshenv1beta1.GroupExtList, error) {
	result := menshenv1beta1.GroupExtList{}
	err := c.restClient.
		Get().
		Resource("groupexts").
		Do().
		Into(&result)

	return &result, err
}
