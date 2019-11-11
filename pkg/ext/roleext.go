package ext

import (
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type RoleExtInterface interface {
	List(opts metav1.ListOptions) (*menshenv1beta1.RoleExtList, error)
	Get(resourceName string) (*menshenv1beta1.RoleExt, error)
}

type roleextClient struct {
	restClient rest.Interface
}

func (c *roleextClient) List(opts metav1.ListOptions) (*menshenv1beta1.RoleExtList, error) {
	result := menshenv1beta1.RoleExtList{}
	err := c.restClient.
		Get().
		Resource("roleexts").
		Do().
		Into(&result)

	return &result, err
}

func (c *roleextClient) Get(resourceName string) (*menshenv1beta1.RoleExt, error) {
	result := menshenv1beta1.RoleExt{}
	err := c.restClient.
		Get().
		Resource("roleexts").
		Name(resourceName).
		Do().
		Into(&result)

	return &result, err
}
