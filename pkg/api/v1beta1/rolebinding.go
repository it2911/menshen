package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type RoleBindingExtInterface interface {
	List(opts metav1.ListOptions) (*RoleBindingExtList, error)
}

type rolebindingextClient struct {
	restClient rest.Interface
}

func (c *rolebindingextClient) List(opts metav1.ListOptions) (*RoleBindingExtList, error) {
	result := RoleBindingExtList{}
	err := c.restClient.
		Get().
		Resource("rolebindings").
		Do().
		Into(&result)

	return &result, err
}
