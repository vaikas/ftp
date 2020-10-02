/*
Copyright 2018 The Knative Authors

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

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type FTPFileEvent struct {
	Name    string
	Size    int64
	ModTime time.Time
}

func receive(ctx context.Context, event cloudevents.Event) error {
	fmt.Printf("Got Event Context: %+v\n", event.Context)
	var data FTPFileEvent
	if err := event.DataAs(&data); err != nil {
		fmt.Printf("Got Data Error: %s\n", err.Error())
	}
	fmt.Printf("Got Data: %+v\n", data)

	fmt.Printf("----------------------------\n")
	return nil
}

func main() {
	ctx := context.TODO()

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	if err := c.StartReceiver(ctx, receive); err != nil {
		log.Fatalf("failed to start receiver: %s", err.Error())
	}

	log.Printf("listening on port %d\n", 8080)
}
