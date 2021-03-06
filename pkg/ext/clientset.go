package ext

import (
	menshenv1beta1 "github.com/it2911/menshen/pkg/api/v1beta1"
	"github.com/it2911/menshen/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

var MenShenClientSet *MenshenV1Beta1Client

type MenshenV1Beta1Interface interface {
	UserExts(namespace string) UserExtInterface
}

type MenshenV1Beta1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*MenshenV1Beta1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &menshenv1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &MenshenV1Beta1Client{restClient: client}, nil
}

func InitClient() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	// BuildConfigFromFlags is a helper function that builds configs from a master url or
	// a kubeconfig filepath.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	utils.HandleErr(err)

	MenShenClientSet, err = NewForConfig(config)
	utils.HandleErr(err)
}

func (c *MenshenV1Beta1Client) UserExts() UserExtInterface {
	return &userextClient{restClient: c.restClient}
}

func (c *MenshenV1Beta1Client) RoleExts() RoleExtInterface {
	return &roleextClient{restClient: c.restClient}
}

func (c *MenshenV1Beta1Client) GroupExts() GroupExtInterface {
	return &groupextClient{restClient: c.restClient}
}

func (c *MenshenV1Beta1Client) RoleBindingExts() RoleBindingExtInterface {
	return &rolebindingextClient{restClient: c.restClient}
}
