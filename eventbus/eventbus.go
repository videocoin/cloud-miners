package eventbus

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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
	logger *logrus.Entry
	mq     *mqmux.WorkerMux
}

func New(c *Config) (*EventBus, error) {
	mq, err := mqmux.NewWorkerMux(c.URI, c.Name)
	if err != nil {
		return nil, err
	}
	if c.Logger != nil {
		mq.Logger = c.Logger
	}

	return &EventBus{
		logger: c.Logger,
		mq:     mq,
	}, nil
}

func (e *EventBus) Start() error {
	err := e.mq.Publisher("streams.status")
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
		span.Tracer().Inject(
			span.Context(),
			opentracing.TextMap,
			mqmux.RMQHeaderCarrier(headers),
		)
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
