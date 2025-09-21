package localobj

import (
	"context"
	"fmt"
)

func NewDefaultStore() (*Store, error) {
	port, err := getAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get available port: %w", err)
	}

	opts := &StoreOptions{
		CmdContext:    defaultMinIOCmdContextWithPort(port),
		ClientOptions: defaultMinIOClientOptions(port),
	}

	return NewStore(opts)
}

func StartDefaultStore(ctx context.Context) (*Store, error) {
	store, err := NewDefaultStore()
	if err != nil {
		return nil, err
	}

	if err := store.Start(ctx); err != nil {
		return nil, err
	}

	if err := store.PollReady(ctx); err != nil {
		store.GracefulShutdown(ctx)
		return nil, err
	}

	return store, nil
}
