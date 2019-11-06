package ext

import (
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type RoleBindingExtInterface interface {
	List(opts metav1.ListOptions) (*menshenv1beta1.RoleBindingExtList, error)
}

type rolebindingextClient struct {
	restClient rest.Interface
}

func (c *rolebindingextClient) List(opts metav1.ListOptions) (*menshenv1beta1.RoleBindingExtList, error) {
	result := menshenv1beta1.RoleBindingExtList{}
	err := c.restClient.
		Get().
		Resource("rolebindings").
		Do().
		Into(&result)

	return &result, err
}
