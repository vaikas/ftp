// Simple abstraction for storing state on a k8s ConfigMap. Very very simple
// and uses a single entry in the ConfigMap.data for storing serialized
// JSON of the generic data that Load/Save uses.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/system"
)

const (
	configdatakey = "configdata"
)

var (
	storeName = "sftp-store"
)

type configstore struct {
	data string
}

type store struct {
	*configmap.UntypedStore
}

type cfgKey struct{}

func NewSFTPStoreConfigFromConfigMap(config *corev1.ConfigMap) (*configstore, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	return newConfigFromMap(config.Data)
}

// newConfigFromMap returns a Config given a map corresponding to a ConfigMap
func newConfigFromMap(cfgMap map[string]string) (*configstore, error) {
	data, ok := cfgMap[configdatakey]
	if !ok {
		return nil, fmt.Errorf("config data not present")
	}
	return &configstore{data}, nil
}

// Initialize ConfigStore. Basically ensures that we can create or get
// the initial configmap. This can fail during startup while Istio
// sidecar is not ready yet for a little while (~seconds), and if the
// ServiceAccount we run as is not configured properly, so we make sure
// things are right before proceeding.
func initConfigStore(ctx context.Context, name string) error {
	//override default store name
	if name != "" {
		storeName = name
	}

	logger := logging.FromContext(ctx)
	logger.Info("Initializing ConfigStore...")

	err := loadConfigMapData(ctx)
	switch {
	case err == nil:
		logger.Info("Config loaded succsesfully")
	case apierrors.IsNotFound(err):
		if err := createEmptyConfigMapData(ctx); err != nil {
			logger.Error("Failed to create empty configmap", zap.Error(err))
			return err
		}
	default:
		logger.Error("Failed to load configmap data:", zap.Error(err))
	}

	return nil
}

// loadConfigMapData loads the ConfigMap and grabs the configmapkey value from
// the map that contains our state.
func loadConfigMapData(ctx context.Context) error {
	cmClient := kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace())
	if err := wait.PollImmediate(1*time.Second, 30*time.Second, func() (bool, error) {
		var err error
		_, err = cmClient.Get(ctx, storeName, metav1.GetOptions{})
		if err == nil || apierrors.IsNotFound(err) {
			return true, err
		}
		logging.FromContext(ctx).Info("Waiting to load config map:", zap.Error(err))
		return false, nil
	}); err != nil {
		return err
	}
	return nil
}

func createEmptyConfigMapData(ctx context.Context) error {
	cmClient := kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace())
	bytes, err := json.Marshal(configdata{})
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
			Name: storeName,
		},
		Data: map[string]string{configdatakey: string(bytes)},
	}
	_, err = cmClient.Create(ctx, cm, metav1.CreateOptions{})
	return err
}

// save takes the value given in, and marshals it into a string
// and saves it into the k8s ConfigMap under the configdatakey.
func (cs *configstore) save(ctx context.Context, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to Marshal", zap.Error(err))
		return err
	}
	cs.data = string(bytes)
	return cs.saveConfigMapData(ctx)
}

// saveConfigMapData saves the ConfigMap with the data from configstore.data
// stored in the configmapkey value of that map.
func (cs *configstore) saveConfigMapData(ctx context.Context) error {
	cmClient := kubeclient.Get(ctx).CoreV1().ConfigMaps(system.Namespace())
	cm, err := cmClient.Get(ctx, storeName, metav1.GetOptions{})
	if err != nil {
		logging.FromContext(ctx).Error("Failed to save CM:", zap.Error(err))
		return err
	}
	cm.Data[configdatakey] = cs.data
	_, err = cmClient.Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

// NewConfigStore creates a new configuration Store.
func NewConfigStore(logger configmap.Logger) *store {
	return &store{
		UntypedStore: configmap.NewUntypedStore(
			"ftpsource",
			logger,
			configmap.Constructors{
				storeName: NewSFTPStoreConfigFromConfigMap,
			},
		),
	}
}

// Load creates a Config for this store.
func (s *store) Load() *configstore {
	return s.UntypedLoad(storeName).(*configstore)
}
