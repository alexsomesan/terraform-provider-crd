package provider

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClients struct {
	config    *rest.Config
	discovery *discovery.DiscoveryClient
}

func NewKubernetesClient() *KubernetesClients {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil)
	clientConfig, err := cc.ClientConfig()
	if err != nil {
		panic(err)
	}
	disClient := discovery.NewDiscoveryClientForConfigOrDie(clientConfig)

	return &KubernetesClients{
		config:    clientConfig,
		discovery: disClient,
	}
}
