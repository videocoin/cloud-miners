package manager

import (
	"github.com/sirupsen/logrus"
	"github.com/videocoin/cloud-miners/datastore"
	"github.com/videocoin/cloud-miners/eventbus"
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
