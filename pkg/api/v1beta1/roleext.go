package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type RoleExtInterface interface {
	List(opts metav1.ListOptions) (*RoleExtList, error)
}

type roleextClient struct {
	restClient rest.Interface
}

func (c *roleextClient) List(opts metav1.ListOptions) (*RoleExtList, error) {
	result := RoleExtList{}
	err := c.restClient.
		Get().
		Resource("roleexts").
		Do().
		Into(&result)

	return &result, err
}
