package ext

import (
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type UserExtInterface interface {
	List(opts metav1.ListOptions) (*menshenv1beta1.UserExtList, error)
}

type userextClient struct {
	restClient rest.Interface
}

func (c *userextClient) List(opts metav1.ListOptions) (*menshenv1beta1.UserExtList, error) {
	result := menshenv1beta1.UserExtList{}
	err := c.restClient.
		Get().
		Resource("userexts").
		Do().
		Into(&result)

	return &result, err
}
