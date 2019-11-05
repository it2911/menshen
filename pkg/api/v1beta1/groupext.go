package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type GroupExtInterface interface {
	List(opts metav1.ListOptions) (*GroupExtList, error)
}

type groupextClient struct {
	restClient rest.Interface
}

func (c *groupextClient) List(opts metav1.ListOptions) (*GroupExtList, error) {
	result := GroupExtList{}
	err := c.restClient.
		Get().
		Resource("groupexts").
		Do().
		Into(&result)

	return &result, err
}
