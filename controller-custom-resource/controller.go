package main

import (
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	clientset "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/clientset/versioned"
	networkscheme "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/clientset/versioned/scheme"
	informers "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/informers/externalversions/samplecrd/v1"
	listers "github.com/justxuewei/k8s_starter/samplecrd/pkg/client/listers/samplecrd/v1"
)

const controllerAgentName = "network-controller"

var (
	// SuccessSynced is used as part of the Event 'reason' when a Network is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a Network
	// is synced successfully
	MessageResourceSynced = "Network synced successfully"
)

// Controller is a controller implementation for Network resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// networkclientset is a clientset for our own API group
	networkclientset clientset.Interface

	networksLister listers.NetworkLister
	networksSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new network controller
func NewController(
	kubeclientset kubernetes.Interface,
	networkclientset clientset.Interface,
	networkInformer informers.NetworkInformer) *Controller {
	// Create event broadcaster
	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(networkscheme.AddToScheme(scheme.Scheme))
	klog.Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})

	controller := &Controller{
		kubeclientset: kubeclientset,
		networkclientset: networkclientset,
		networksLister: networkInformer.Lister(),
	}

	return controller
}
