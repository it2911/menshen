package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type UserExtInterface interface {
	List(opts metav1.ListOptions) (*UserExtList, error)
}

type userextClient struct {
	restClient rest.Interface
}

func (c *userextClient) List(opts metav1.ListOptions) (*UserExtList, error) {
	result := UserExtList{}
	err := c.restClient.
		Get().
		Resource("userexts").
		Do().
		Into(&result)

	return &result, err
}
