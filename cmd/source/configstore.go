// Simple abstraction for storing state on a k8s ConfigMap. Very very simple
// and uses a single entry in the ConfigMap.data for storing serialized
// JSON of the generic data that Load/Save uses.
package main

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/typed/core/v1"

	"knative.dev/pkg/logging"
)

const (
	configdatakey = "configdata"
)

type configstore struct {
	cmClient  v1.ConfigMapInterface
	name      string
	namespace string
	data      string
}

func NewConfigStore(name string, namespace string, cmClient v1.ConfigMapInterface) *configstore {
	return &configstore{name: name, namespace: namespace, cmClient: cmClient}
}

// Initialize ConfigStore. Basically ensures that we can create or get
// the initial configmap. This can fail during startup while Istio
// sidecar is not ready yet for a little while (~seconds), and if the
// ServiceAccount we run as is not configured properly, so we make sure
// things are right before proceeding.
func (cs *configstore) Init(ctx context.Context, value interface{}) error {
	logger := logging.FromContext(ctx)
	logger.Info("Initializing ConfigStore...")

	err := cs.loadConfigMapData(ctx)
	switch {
	case err == nil:
		logger.Info("Config loaded succsesfully")
	case apierrors.IsNotFound(err):
		if err := cs.createEmptyConfigMapData(ctx, value); err != nil {
			logger.Error("Failed to create empty configmap", zap.Error(err))
			return err
		}
	default:
		logger.Error("Failed to load configmap data:", zap.Error(err))
	}

	return nil
}

func (cs *configstore) createEmptyConfigMapData(ctx context.Context, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to Marshal:", zap.Error(err))
		return err
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cs.name,
			Namespace: cs.namespace,
		},
		Data: map[string]string{configdatakey: string(bytes)},
	}
	_, err = cs.cmClient.Create(ctx, cm, metav1.CreateOptions{})
	return err
}

// loadConfigMapData loads the ConfigMap and grabs the configmapkey value from
// the map that contains our state.
func (cs *configstore) loadConfigMapData(ctx context.Context) error {
	var cm *corev1.ConfigMap
	if err := wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
		var err error
		cm, err = cs.cmClient.Get(ctx, cs.name, metav1.GetOptions{})
		return err == nil || apierrors.IsNotFound(err), err
	}); err != nil {
		return err
	}
	cs.data = cm.Data[configdatakey]
	return nil
}

// saveConfigMapData saves the ConfigMap with the data from configstore.data
// stored in the configmapkey value of that map.
func (cs *configstore) saveConfigMapData(ctx context.Context) error {
	cm, err := cs.cmClient.Get(ctx, cs.name, metav1.GetOptions{})
	if err != nil {
		logging.FromContext(ctx).Error("Failed to save CM:", zap.Error(err))
		return err
	}
	cm.Data[configdatakey] = cs.data
	_, err = cs.cmClient.Update(ctx, cm, metav1.UpdateOptions{})
	return err

}

// Load fetches the ConfigMap from k8s and unmarshals the data found
// in the configdatakey type as specified by value.
func (cs *configstore) Load(ctx context.Context, value interface{}) error {
	err := cs.loadConfigMapData(ctx)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(cs.data), value)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to Unmarshal", zap.Error(err))
		return err
	}
	return nil
}

// save takes the value given in, and marshals it into a string
// and saves it into the k8s ConfigMap under the configdatakey.
func (cs *configstore) Save(ctx context.Context, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to Marshal", zap.Error(err))
		return err
	}
	cs.data = string(bytes)
	return cs.saveConfigMapData(ctx)
}
