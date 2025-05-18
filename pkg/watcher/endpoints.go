package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/hanapedia/maglseven/pkg/maglev"
	"github.com/hanapedia/maglseven/pkg/util"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type EndpointSlicesWatcher struct {
	Namespace string
	Service   string
}

func (w *EndpointSlicesWatcher) Watch(ctx context.Context, updates chan<- []maglev.Backend) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Minute,
		informers.WithNamespace(w.Namespace),
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = fmt.Sprintf("kubernetes.io/service-name=%s", w.Service)
		}),
	)

	informer := factory.Discovery().V1().EndpointSlices().Informer()

	var lastHash string

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			w.processUpdate(informer, updates, &lastHash)
		},
		UpdateFunc: func(oldObj, newObj any) {
			w.processUpdate(informer, updates, &lastHash)
		},
		DeleteFunc: func(obj any) {
			w.processUpdate(informer, updates, &lastHash)
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	factory.Start(stopCh)
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	<-ctx.Done()
	return ctx.Err()
}

func (w *EndpointSlicesWatcher) processUpdate(informer cache.SharedIndexInformer, updates chan<- []maglev.Backend, lastHash *string) {
	slices := informer.GetStore().List()

	var backends []maglev.Backend
	for _, obj := range slices {
		es, ok := obj.(*discoveryv1.EndpointSlice)
		if !ok {
			continue
		}

		for _, endpoint := range es.Endpoints {
			if endpoint.Conditions.Ready != nil && *endpoint.Conditions.Ready {
				for _, addr := range endpoint.Addresses {
					backends = append(backends, maglev.Backend{ID: addr})
				}
			}
		}
	}

	newHash := util.HashBackends(backends)
	if newHash != *lastHash {
		*lastHash = newHash
		select {
		case updates <- backends:
		default:
		}
	}
}
