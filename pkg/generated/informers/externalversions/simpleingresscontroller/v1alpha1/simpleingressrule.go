/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	simpleingresscontrollerv1alpha1 "github.com/nimrodoron/simple-ingress-controller/pkg/apis/simpleingresscontroller/v1alpha1"
	versioned "github.com/nimrodoron/simple-ingress-controller/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/nimrodoron/simple-ingress-controller/pkg/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/nimrodoron/simple-ingress-controller/pkg/generated/listers/simpleingresscontroller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// SimpleIngressRuleInformer provides access to a shared informer and lister for
// SimpleIngressRules.
type SimpleIngressRuleInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.SimpleIngressRuleLister
}

type simpleIngressRuleInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewSimpleIngressRuleInformer constructs a new informer for SimpleIngressRule type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewSimpleIngressRuleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredSimpleIngressRuleInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredSimpleIngressRuleInformer constructs a new informer for SimpleIngressRule type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredSimpleIngressRuleInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecontrollerV1alpha1().SimpleIngressRules(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecontrollerV1alpha1().SimpleIngressRules(namespace).Watch(options)
			},
		},
		&simpleingresscontrollerv1alpha1.SimpleIngressRule{},
		resyncPeriod,
		indexers,
	)
}

func (f *simpleIngressRuleInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredSimpleIngressRuleInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *simpleIngressRuleInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&simpleingresscontrollerv1alpha1.SimpleIngressRule{}, f.defaultInformer)
}

func (f *simpleIngressRuleInformer) Lister() v1alpha1.SimpleIngressRuleLister {
	return v1alpha1.NewSimpleIngressRuleLister(f.Informer().GetIndexer())
}
