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
	"context"
	time "time"

	raihankhangithubiov1alpha1 "github.com/raihankhan/httpApiServer-controller/pkg/apis/raihankhan.github.io/v1alpha1"
	versioned "github.com/raihankhan/httpApiServer-controller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/raihankhan/httpApiServer-controller/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/raihankhan/httpApiServer-controller/pkg/client/listers/raihankhan.github.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ApiserverInformer provides access to a shared informer and lister for
// Apiservers.
type ApiserverInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ApiserverLister
}

type apiserverInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewApiserverInformer constructs a new informer for Apiserver type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewApiserverInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredApiserverInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredApiserverInformer constructs a new informer for Apiserver type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredApiserverInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.RaihankhanV1alpha1().Apiservers(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.RaihankhanV1alpha1().Apiservers(namespace).Watch(context.TODO(), options)
			},
		},
		&raihankhangithubiov1alpha1.Apiserver{},
		resyncPeriod,
		indexers,
	)
}

func (f *apiserverInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredApiserverInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *apiserverInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&raihankhangithubiov1alpha1.Apiserver{}, f.defaultInformer)
}

func (f *apiserverInformer) Lister() v1alpha1.ApiserverLister {
	return v1alpha1.NewApiserverLister(f.Informer().GetIndexer())
}
