package manager

import (
	"github.com/sirupsen/logrus"
	emitterv1 "github.com/videocoin/cloud-api/emitter/v1"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-miners/eventbus"
	"github.com/videocoin/cloud-pkg/grpcutil"
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

func WithEventBus(eb *eventbus.EventBus) Option {
	return func(m *Manager) error {
		m.eb = eb
		return nil
	}
}

func WithEmitterServiceClient(addr string) Option {
	return func(m *Manager) error {
		conn, err := grpcutil.Connect(addr, m.logger.WithField("system", "emitter"))
		if err != nil {
			return err
		}
		m.emitter = emitterv1.NewEmitterServiceClient(conn)
		return nil
	}
}
