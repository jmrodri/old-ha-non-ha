package main

import (
	"fmt"
	"os"
	"time"

	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformer "github.com/openshift/client-go/config/informers/externalversions"
	configv1lister "github.com/openshift/client-go/config/listers/config/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var infrastructureLister configv1lister.InfrastructureLister

func main() {
	fmt.Println("entered main")

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	mergedConfig, err := loadingRules.Load()
	if err != nil {
		fmt.Printf("problem loading kubeconfig: %v\n", err)
		os.Exit(-3)
	}

	fmt.Println("creating newdefaultclientconfig")
	cfg := clientcmd.NewDefaultClientConfig(*mergedConfig, nil)
	cc, err := cfg.ClientConfig()
	if err != nil {
		fmt.Printf("problem using ClientConfig: %v\n", err)
		os.Exit(-1)
	}

	fmt.Println("creating new auth operator client")

	configClient, err := configclient.NewForConfig(cc)
	if err != nil {
		fmt.Printf("problem creating new client from config: %v\n", err)
		os.Exit(-2)
	}

	operatorConfigInformer := configinformer.NewSharedInformerFactoryWithOptions(configClient, 20*time.Minute)
	infrastructureLister = operatorConfigInformer.Config().V1().Infrastructures().Lister()

	infraConfig, err := infrastructureLister.Get("cluster")
	if err != nil {
		fmt.Printf("problem getting infrastructure config: %v\n", err)
		os.Exit(-4)
	}

	fmt.Printf("%v\n", infraConfig)
}
