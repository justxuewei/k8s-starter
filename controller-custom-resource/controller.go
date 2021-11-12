package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	samplecrdv1 "github.com/justxuewei/k8s_starter/samplecrd/pkg/apis/samplecrd/v1"
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
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset: kubeclientset,
		networkclientset: networkclientset,
		networksLister: networkInformer.Lister(),
		networksSynced: networkInformer.Informer().HasSynced,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Networks"),
		recorder: recorder,
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Network resources change
	networkInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueNetwork,
		UpdateFunc: func(old, new interface{}) {
			oldNetwork := old.(*samplecrdv1.Network)
			newNetwork := new.(*samplecrdv1.Network)
			if oldNetwork.ResourceVersion == newNetwork.ResourceVersion {
				return
			}
			controller.enqueueNetwork(new)
		},
		DeleteFunc: controller.enqueueNetworkForDelete,
	})

	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	klog.Info("Starting Network control loop")

	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.networksSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	for i:=0; i<threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	klog.Info("Started workers")

	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workerqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		var (
			key string
			ok bool
		)

		// We expect strings to come off the workqueue. There are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the item in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		// Run the syncHandler, passing it the namespace/name string of the
		// Network resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': '%s'", key, err.Error())
		}

		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)

		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: '%s'", key))
		return nil
	}

	// Get the Network resource with this namespace/name
	network, err := c.networksLister.Networks(namespace).Get(name)
	if err != nil {
		// The Network resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			klog.Warningf("Network: %s/%s does not exist in local cache, will delete it from Neutron...",
				namespace, name)
			klog.Infof("[Neutron] Deleting network: %s/%s...", namespace, name)

			// FIXME: call Neutron API to delete this network by name.
			// neutron.Delete(namespace, name)

			return nil
		}

		runtime.HandleError(fmt.Errorf("failed to list network by: %s/%s", namespace, name))

		return err
	}

	klog.Infof("[Neutron] Try to process network: %#")

	c.recorder.Eventf(network, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)

	return nil
}

func (c *Controller) enqueueNetwork(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) enqueueNetworkForDelete(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}
