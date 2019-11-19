package main

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	serviceInformer "k8s.io/client-go/informers/core/v1"
	servicelisters "k8s.io/client-go/listers/core/v1"

	"github.com/nimrodoron/simple-ingress-controller/pkg/apis/simpleingresscontroller/v1alpha1"
	controllerlientset "github.com/nimrodoron/simple-ingress-controller/pkg/generated/clientset/versioned"
	informers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/informers/externalversions/simpleingresscontroller/v1alpha1"
	listers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/listers/simpleingresscontroller/v1alpha1"

	reverseproxy "github.com/nimrodoron/simple-ingress-controller/pkg/server"
)

// Controller object
type Controller struct {
	kubeclientset           kubernetes.Interface
	controllerclientset     controllerlientset.Interface
	simpleIngressRuleLister listers.SimpleIngressRuleLister
	simpleIngressRuleSynced cache.InformerSynced
	serviceLister           servicelisters.ServiceLister
	serviceSynced           cache.InformerSynced
	rulesCh                 chan<- *reverseproxy.ProxyRuleOperation
	workqueue               workqueue.RateLimitingInterface
}

func NewController(
	kubeclientset kubernetes.Interface,
	controllerclientset controllerlientset.Interface,
	simpleIngressRuleInformer informers.SimpleIngressRuleInformer,
	serviceInformer serviceInformer.ServiceInformer,
	ruleCh chan<- *reverseproxy.ProxyRuleOperation) *Controller {

	controller := &Controller{
		kubeclientset:           kubeclientset,
		controllerclientset:     controllerclientset,
		simpleIngressRuleLister: simpleIngressRuleInformer.Lister(),
		simpleIngressRuleSynced: simpleIngressRuleInformer.Informer().HasSynced,
		serviceLister:           serviceInformer.Lister(),
		serviceSynced:           serviceInformer.Informer().HasSynced,
		rulesCh:                 ruleCh,
		workqueue:               workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SimpleIngressRules"),
	}

	simpleIngressRuleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*v1alpha1.SimpleIngressRule)
			oldDepl := old.(*v1alpha1.SimpleIngressRule)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

func (c *Controller) handleObject(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Simple Ingress controller")

	if ok := cache.WaitForCacheSync(stopCh, c.simpleIngressRuleSynced) && cache.WaitForCacheSync(stopCh, c.serviceSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
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

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		} else {
			simpleIngressRule, err2 := c.simpleIngressRuleLister.SimpleIngressRules(namespace).Get(name)
			if err2 == nil {
				rulesMap := make(map[string]string)
				for _, rule := range simpleIngressRule.Spec.Rules {
					rulesMap[rule.Path] = rule.Service.Name
				}
				c.rulesCh <- &reverseproxy.ProxyRuleOperation{
					Name:      simpleIngressRule.Name,
					Rules:     rulesMap,
					Operation: reverseproxy.ADD,
				}

			} else {
				c.rulesCh <- &reverseproxy.ProxyRuleOperation{
					Name:      name,
					Rules:     nil,
					Operation: reverseproxy.REMOVE,
				}
			}
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

func (c *Controller) GetService(name string) (*reverseproxy.Service, error) {
	service, err := c.serviceLister.Services("default").Get(name)
	if err != nil {
		return nil, err
	} else {
		return &reverseproxy.Service{
			Address: "http://" + service.Spec.ClusterIP,
			Port:    fmt.Sprint(service.Spec.Ports[0].Port),
		}, nil
	}
}
