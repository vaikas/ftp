package main

import (
	"context"
	"os"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"knative.dev/pkg/logging"
)

type publisher struct {
	ceClient     cloudevents.Client
	sourceServer string
}

type FTPFileEvent struct {
	Name    string
	Size    int64
	ModTime time.Time
}

func (p *publisher) postMessage(ctx context.Context, fileEntry os.FileInfo) error {
	logger := logging.FromContext(ctx)
	d := FTPFileEvent{Name: fileEntry.Name(), Size: fileEntry.Size(), ModTime: fileEntry.ModTime()}
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	id, _ := uuid.NewUUID()
	event.SetID(id.String())
	event.SetTime(time.Now())
	event.SetType(event_type)
	//adding slashes to avoid parse url errors
	event.SetSource("//" + p.sourceServer)

	if err := event.SetData(cloudevents.ApplicationJSON, d); err != nil {
		logger.Error("Error setting data ", zap.Error(err))
		return err
	}

	logger.Info("posting message to sink")

	if err := p.ceClient.Send(ctx, event); !cloudevents.IsACK(err) {
		logger.Error("failed to send cloudevent", zap.Error(err))
		return err
	}
	return nil
}
