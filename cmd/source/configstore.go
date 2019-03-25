// Simple abstraction for storing state on a k8s ConfigMap. Very very simple
// and uses a single entry in the ConfigMap.data for storing serialized
// JSON of the generic data that Load/Save uses.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"time"
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
func (cs *configstore) Init(value interface{}) error {
	fmt.Println("Initializing ConfigStore...")

	for i := 0; i < 10; i++ {
		err := cs.loadConfigMapData()
		if apierrors.IsNotFound(err) {
			fmt.Printf("No config found, creating empty")
			err = cs.createEmptyConfigMapData(value)
			if err != nil {
				fmt.Printf("Failed to create empty configmap: %s\n", err)
			} else {
				fmt.Println("Empty config created successfully")
				err = cs.loadConfigMapData()
				if err == nil {
					fmt.Println("Config loaded succsesfully")
					return nil
				} else {
					fmt.Printf("Failed to load configmap data: %s\n", err)
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	return errors.New("Timed out while trying to create configstore")
}

func (cs *configstore) createEmptyConfigMapData(value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("Failed to Marshal: %s", err)
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
	_, err = cs.cmClient.Create(cm)
	return err
}

// loadConfigMapData loads the ConfigMap and grabs the configmapkey value from
// the map that contains our state.
func (cs *configstore) loadConfigMapData() error {
	cm, err := cs.cmClient.Get(cs.name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Failed to get CM: %s\n", err)
		return err
	}
	cs.data = cm.Data[configdatakey]
	return nil
}

// saveConfigMapData saves the ConfigMap with the data from configstore.data
// stored in the configmapkey value of that map.
func (cs *configstore) saveConfigMapData() error {
	cm, err := cs.cmClient.Get(cs.name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("Failed to get CM: %s\n", err)
		return err
	}
	cm.Data[configdatakey] = cs.data
	_, err = cs.cmClient.Update(cm)
	return err

}

// Load fetches the ConfigMap from k8s and unmarshals the data found
// in the configdatakey type as specified by value.
func (cs *configstore) Load(value interface{}) error {
	err := cs.loadConfigMapData()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(cs.data), value)
	if err != nil {
		fmt.Printf("Failed to Unmarshal: %s", err)
		return err
	}
	return nil
}

// save takes the value given in, and marshals it into a string
// and saves it into the k8s ConfigMap under the configdatakey.
func (cs *configstore) Save(value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("Failed to Marshal: %s\n", err)
		return err
	}
	cs.data = string(bytes)
	return cs.saveConfigMapData()
}
