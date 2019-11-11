package ext

import (
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type RoleBindingExtInterface interface {
	List(opts metav1.ListOptions) (*menshenv1beta1.RoleBindingExtList, error)
	Get(resourceName string) (*menshenv1beta1.RoleBindingExt, error)
}

type rolebindingextClient struct {
	restClient rest.Interface
}

func (c *rolebindingextClient) Get(resourceName string) (*menshenv1beta1.RoleBindingExt, error) {
	result := menshenv1beta1.RoleBindingExt{}
	err := c.restClient.
		Get().
		Resource("rolebindingexts").
		Name(resourceName).
		Do().
		Into(&result)

	return &result, err
}

func (c *rolebindingextClient) List(opts metav1.ListOptions) (*menshenv1beta1.RoleBindingExtList, error) {
	result := menshenv1beta1.RoleBindingExtList{}
	err := c.restClient.
		Get().
		Resource("rolebindingexts").
		Do().
		Into(&result)

	return &result, err
}
