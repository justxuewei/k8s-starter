package main

import (
	"flag"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clientset "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/clientset/versioned"
	informers "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/informers/externalversions"
	"github.com/justxuewei/k8s_starter/samplecrd/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "",
		"The address of the Kubernetes API server. Overrides any value in kubeconfig. " +
		"Only required if out-of-cluster.")
}

func main() {
	flag.Parse()

	// set up signals, so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// kubernetes client
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	// network client
	networkClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building network clientset: %s", err.Error())
	}

	// get network informer factory
	networkInformerFactory := informers.NewSharedInformerFactory(networkClient, time.Second * 30)

	controller := NewController(kubeClient, networkClient, networkInformerFactory.Samplecrd().V1().Networks())

	// TODO(justxuewei): What does this line of code do?
	go networkInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
