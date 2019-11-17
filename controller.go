package main

import (
	"fmt"
	"k8s.io/klog"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	controllerlientset "github.com/nimrodoron/simple-ingress-controller/pkg/generated/clientset/versioned"
	informers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/informers/externalversions"
	listers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/listers/simpleingresscontroller/v1alpha1"
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

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Simple Ingress controller")

	c.simpleIngressRuleInformerFactory.Start(stopCh)

	if ok := cache.WaitForCacheSync(stopCh, c.simpleIngressRuleSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	return nil
}
