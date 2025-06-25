package provider

import (
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/openapi"
	"k8s.io/client-go/openapi3"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClients struct {
	Config        *rest.Config
	Discovery     *discovery.DiscoveryClient
	APIextensions *apiextensionsclientset.Clientset
	Openapi       openapi3.Root
}

func NewKubernetesClient() *KubernetesClients {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil)
	clientConfig, err := cc.ClientConfig()
	if err != nil {
		panic(err)
	}
	disClient := discovery.NewDiscoveryClientForConfigOrDie(clientConfig)
	oapi := openapi3.NewRoot(openapi.NewClient(disClient.RESTClient()))

	return &KubernetesClients{
		Config:        clientConfig,
		Discovery:     disClient,
		APIextensions: apiextensionsclientset.NewForConfigOrDie(clientConfig),
		Openapi:       oapi,
	}
}
