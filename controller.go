package main

/*import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Controller object
type Controller struct {
	kubeclientset    kubernetes.Interface
	workqueue        workqueue.RateLimitingInterface
	ingressInformer     cache.SharedInformer

}

func NewController(kubeclientset    kubernetes.Interface,
	) *Controller {
	cache.ListWatch{}
	controller := &Controller{
		kubeclientset:     kubeclientset,
		workqueue:         workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		ingressInformer:   cache.SharedInformer(&cache.ListWatch{
								ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
									return kubeclientset.(meta_v1.NamespaceAll).List(options)
								},
								WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
									return client.CoreV1().Pods(meta_v1.NamespaceAll).Watch(options)
								}
		})
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		foosLister:        fooInformer.Lister(),
		foosSynced:        fooInformer.Informer().HasSynced,

		recorder:          recorder,
	}
}*/
