package eventbus

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	v1 "github.com/videocoin/cloud-api/miners/v1"
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
	err := e.mq.Publisher("miners.events")
	if err != nil {
		return err
	}

	return e.mq.Run()
}

func (e *EventBus) Stop() error {
	return e.mq.Close()
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
