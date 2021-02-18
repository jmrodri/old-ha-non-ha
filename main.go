package main

import (
	"context"
	"fmt"
	"os"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	configinformer "github.com/openshift/client-go/config/informers/externalversions"
	configv1lister "github.com/openshift/client-go/config/listers/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var infrastructureLister configv1lister.InfrastructureLister

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
}

func GetInfraViaK8S(cc *restclient.Config) error {
	// Create a new config client
	configClient, err := configclient.NewForConfig(cc)
	if err != nil {
		return err
	}
	operatorConfigInformer := configinformer.NewSharedInformerFactoryWithOptions(configClient, 2*time.Second)
	infrastructureLister = operatorConfigInformer.Config().V1().Infrastructures().Lister()
	infraConfig, err := configClient.ConfigV1().Infrastructures().Get(context.Background(), "cluster", metav1.GetOptions{})
	if err != nil {
		return err
	}

	// fmt.Printf("%v\n", infraConfig)
	fmt.Printf("%v\n", infraConfig.Status.ControlPlaneTopology)
	fmt.Printf("%v\n", infraConfig.Status.InfrastructureTopology)
	return nil
}

func GetInfraViaControllerRuntime(cc *restclient.Config) error {
	rm, err := apiutil.NewDynamicRESTMapper(cc)
	if err != nil {
		return err
	}

	crClient, err := crclient.New(cc, crclient.Options{
		Scheme: scheme,
		Mapper: rm,
	})
	if err != nil {
		return err
	}

	// Simple query
	nn := types.NamespacedName{
		Name: "cluster",
	}
	infraConfig := &configv1.Infrastructure{}
	err = crClient.Get(context.Background(), nn, infraConfig)
	if err != nil {
		return err
	}
	fmt.Printf("using crclient: %v\n", infraConfig.Status.ControlPlaneTopology)
	fmt.Printf("using crclient: %v\n", infraConfig.Status.InfrastructureTopology)

	return nil
}

func main() {
	fmt.Println("entered main")

	// Load the config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	mergedConfig, err := loadingRules.Load()
	if err != nil {
		fmt.Printf("problem loading kubeconfig: %v\n", err)
		os.Exit(-3)
	}
	fmt.Printf("clusters: %v\n", mergedConfig.Clusters)

	cfg := clientcmd.NewDefaultClientConfig(*mergedConfig, nil)
	cc, err := cfg.ClientConfig()
	if err != nil {
		fmt.Printf("problem using ClientConfig: %v\n", err)
		os.Exit(-1)
	}

	// Get the Infra using straight K8S
	err = GetInfraViaK8S(cc)
	if err != nil {
		fmt.Printf("problem getting via straight k8s: %v\n", err)
		os.Exit(-4)
	}

	// Get the Infra using controller runtime
	err = GetInfraViaControllerRuntime(cc)
	if err != nil {
		fmt.Printf("problem getting via controller runtime: %v\n", err)
		os.Exit(-2)
	}
}
