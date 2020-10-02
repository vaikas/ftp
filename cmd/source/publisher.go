package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type publisher struct {
	logger       *zap.Logger
	ceClient     cloudevents.Client
	sourceServer string
}

type FTPFileEvent struct {
	Name    string
	Size    int64
	ModTime time.Time
}

func (p *publisher) postMessage(ctx context.Context, fileEntry os.FileInfo) error {
	d := FTPFileEvent{Name: fileEntry.Name(), Size: fileEntry.Size(), ModTime: fileEntry.ModTime()}
	event := cloudevents.NewEvent(cloudevents.VersionV1)

	id, _ := uuid.NewUUID()
	event.SetID(id.String())
	event.SetTime(time.Now())
	event.SetType(event_type)
	event.SetSource(fmt.Sprintf("%s%s", p.sourceServer, fileEntry.Name()))

	if err := event.SetData(cloudevents.ApplicationJSON, d); err != nil {
		p.logger.Error("Error setting data ", zap.Error(err))
		return err
	}

	if err := p.ceClient.Send(ctx, event); !cloudevents.IsACK(err) {
		p.logger.Error("failed to send gql subscription cloudevent", zap.Error(err))
		return err
	}
	return nil
}
