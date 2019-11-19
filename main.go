package main

import (
	"flag"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	clientset "github.com/nimrodoron/simple-ingress-controller/pkg/generated/clientset/versioned"
	informers "github.com/nimrodoron/simple-ingress-controller/pkg/generated/informers/externalversions"
	"github.com/nimrodoron/simple-ingress-controller/pkg/signals"

	reverseproxy "github.com/nimrodoron/simple-ingress-controller/pkg/server"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	rulesCh := make(chan *reverseproxy.ProxyRuleOperation)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	controllerClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	simpleIngressRuleInformerFactory := informers.NewSharedInformerFactory(controllerClient, time.Second*30)
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	controller := NewController(kubeClient, controllerClient, simpleIngressRuleInformerFactory.Samplecontroller().V1alpha1().SimpleIngressRules(),
		kubeInformerFactory.Core().V1().Services(), rulesCh)

	simpleIngressRuleInformerFactory.Start(stopCh)
	kubeInformerFactory.Start(stopCh)

	proxy := reverseproxy.NewReverseProxy(8080, controller)
	go proxy.Run(rulesCh)

	if err = controller.Run(stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
