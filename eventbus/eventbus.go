package eventbus

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	dispatcherv1 "github.com/videocoin/cloud-api/dispatcher/v1"
	v1 "github.com/videocoin/cloud-api/miners/v1"
	privatev1 "github.com/videocoin/cloud-api/streams/private/v1"
	streamsv1 "github.com/videocoin/cloud-api/streams/v1"
	"github.com/videocoin/cloud-pkg/mqmux"
)

type Config struct {
	Logger *logrus.Entry
	URI    string
	Name   string
}

type EventBus struct {
	mq *mqmux.WorkerMux
}

func New(c *Config) (*EventBus, error) {
	mq, err := mqmux.NewWorkerMux(c.URI, c.Name)
	if err != nil {
		return nil, err
	}

	return &EventBus{
		mq: mq,
	}, nil
}

func (e *EventBus) Start() error {
	err := e.mq.Publisher("streams.status")
	if err != nil {
		return err
	}

	err = e.mq.Publisher("tasks.events")
	if err != nil {
		return err
	}

	err = e.mq.Publisher("miners.events")
	if err != nil {
		return err
	}

	return e.mq.Run()
}

func (e *EventBus) Stop() error {
	return e.mq.Close()
}

func (e *EventBus) EmitUpdateStreamStatus(ctx context.Context, id string, status streamsv1.StreamStatus) error {
	headers := make(amqp.Table)

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "miners")
		err := span.Tracer().Inject(
			span.Context(),
			opentracing.TextMap,
			mqmux.RMQHeaderCarrier(headers),
		)
		if err != nil {
			return err
		}
	}

	event := &privatev1.Event{
		Type:     privatev1.EventTypeUpdateStatus,
		StreamID: id,
		Status:   status,
	}
	err := e.mq.PublishX("streams.status", event, headers)
	if err != nil {
		return err
	}
	return nil
}

func (e *EventBus) EmitUpdateTaskStatus(ctx context.Context, id string, status dispatcherv1.TaskStatus) error {
	headers := make(amqp.Table)

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "miners")
		err := span.Tracer().Inject(
			span.Context(),
			opentracing.TextMap,
			mqmux.RMQHeaderCarrier(headers),
		)
		if err != nil {
			return err
		}
	}

	event := &dispatcherv1.Event{
		Type:   dispatcherv1.EventTypeUpdateStatus,
		TaskID: id,
		Status: status,
	}
	err := e.mq.PublishX("tasks.events", event, headers)
	if err != nil {
		return err
	}
	return nil
}

func (e *EventBus) EmitAssignMinerAddress(ctx context.Context, userID, address string) error {
	headers := make(amqp.Table)

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		ext.SpanKindRPCServer.Set(span)
		ext.Component.Set(span, "miners")
		err := span.Tracer().Inject(
			span.Context(),
			opentracing.TextMap,
			mqmux.RMQHeaderCarrier(headers),
		)
		if err != nil {
			return err
		}
	}

	event := &v1.Event{
		Type:    v1.EventTypeAssignMinerAddress,
		UserID:  userID,
		Address: address,
	}
	err := e.mq.PublishX("miners.events", event, headers)
	if err != nil {
		return err
	}
	return nil
}
