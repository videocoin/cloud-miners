package manager

import (
	"time"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpclogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpctracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Option func(*Manager) error

func WithLogger(logger *logrus.Entry) Option {
	return func(m *Manager) error {
		m.logger = logger
		return nil
	}
}

func WithDatastore(ds *datastore.Datastore) Option {
	return func(m *Manager) error {
		m.ds = ds
		return nil
	}
}

func WithEmitterServiceClient(addr string) Option {
	return func(m *Manager) error {
		opts := []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(
				grpcmiddleware.ChainUnaryClient(
					grpctracing.UnaryClientInterceptor(grpctracing.WithTracer(opentracing.GlobalTracer())),
					grpcprometheus.UnaryClientInterceptor,
					grpclogrus.UnaryClientInterceptor(m.logger.WithField("system", "emitter")),
				),
			),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Second * 10,
				Timeout:             time.Second * 10,
				PermitWithoutStream: true,
			}),
		}
		conn, err := grpc.Dial(addr, opts...)
		if err != nil {
			return err
		}
		m.emitter = emitterv1.NewEmitterServiceClient(conn)
		return nil
	}
}
