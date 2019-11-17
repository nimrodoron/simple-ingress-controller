package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	controllerlientset "github.com/nimrodoron/simple-ingress-controller/pkg/generated/clientset/versioned"
	informers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/informers/externalversions"
	listers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/listers/simpleingresscontroller/v1alpha1"

	reverseproxy "github.com/nimrodoron/simple-ingress-controller/pkg/server"
)

// Controller object
type Controller struct {
	kubeclientset                    kubernetes.Interface
	controllerclientset              controllerlientset.Interface
	simpleIngressRuleLister          listers.SimpleIngressRuleLister
	simpleIngressRuleSynced          cache.InformerSynced
	workqueue                        workqueue.RateLimitingInterface
	simpleIngressRuleInformerFactory informers.SharedInformerFactory
}

func NewController(
	kubeclientset kubernetes.Interface,
	controllerclientset controllerlientset.Interface) *Controller {

	simpleIngressRuleInformerFactory := informers.NewSharedInformerFactory(controllerclientset, time.Second*30)
	simpleIngressRuleInformer := simpleIngressRuleInformerFactory.Samplecontroller().V1alpha1().SimpleIngressRules()

	//simpleIngressRuleInformer.Informer()
	controller := &Controller{
		kubeclientset:                    kubeclientset,
		controllerclientset:              controllerclientset,
		simpleIngressRuleLister:          simpleIngressRuleInformer.Lister(),
		simpleIngressRuleSynced:          simpleIngressRuleInformer.Informer().HasSynced,
		workqueue:                        workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SimpleIngressRules"),
		simpleIngressRuleInformerFactory: simpleIngressRuleInformerFactory,
	}

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}, ruleCh chan<- *reverseproxy.ProxyRuleOperation) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Simple Ingress controller")

	c.simpleIngressRuleInformerFactory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.simpleIngressRuleSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	simpleIngressRules, err := c.simpleIngressRuleLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to retreive list from cache")
	}
	for _, simpleIngressRule := range simpleIngressRules {
		for _, rule := range simpleIngressRule.Spec.Rules {
			ruleCh <- &reverseproxy.ProxyRuleOperation{
				Path:        rule.Path,
				ServiceName: rule.Service.Name,
				Operation:   reverseproxy.ADD,
			}
		}
	}

	klog.Info("Starting workers")
	go wait.Until(c.runWorker, time.Second, stopCh)

	<-stopCh
	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		/*// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.*/
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
