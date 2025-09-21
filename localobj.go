// Package localobj provides a testing library for spawning and managing local object
// store instances. It supports pluggable S3-compatible stores with MinIO as the default.
package localobj

import (
	"context"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	defaultPollTimeout  = 30 * time.Second
	defaultPollInterval = 100 * time.Millisecond
)

var (
	errStoreAlreadyStarted = fmt.Errorf("store already started")
	errStoreNotStarted     = fmt.Errorf("store not started")
)

type CmdContext func(ctx context.Context) (*exec.Cmd, error)

type StoreOptions struct {
	CmdContext    CmdContext
	ClientOptions func(o *s3.Options)
}

type Store struct {
	opts *StoreOptions
	cmd  *exec.Cmd
}

func (s *Store) Start(ctx context.Context) error {
	if s.cmd != nil {
		return errStoreAlreadyStarted
	}

	var err error
	s.cmd, err = s.opts.CmdContext(ctx)
	if err != nil {
		return err
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start store process: %w", err)
	}

	return nil
}

func (s *Store) PollReady(ctx context.Context) error {
	if s.cmd == nil {
		return errStoreNotStarted
	}

	interval := defaultPollInterval
	deadline := time.Now().Add(defaultPollTimeout)

	for time.Now().Before(deadline) {
		client, err := s.NewClient()
		if err != nil {
			time.Sleep(interval)
			continue
		}

		_, err = client.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err == nil {
			return nil
		}

		time.Sleep(interval)
		if interval < time.Second {
			interval *= 2
		}
	}

	return fmt.Errorf("store did not become ready within %v", defaultPollTimeout)
}

func (s *Store) NewClient() (*s3.Client, error) {
	if s.cmd == nil {
		return nil, errStoreNotStarted
	}

	return s3.New(s3.Options{}, s.opts.ClientOptions), nil
}

func (s *Store) GracefulShutdown(ctx context.Context) error {
	if s.cmd == nil {
		return errStoreNotStarted
	}

	if err := s.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to store process: %w", err)
	}

	err := s.cmd.Wait()
	if err != nil {
		return fmt.Errorf("error waiting for store process to exit: %w", err)
	}
	s.cmd = nil

	return nil
}

func NewStore(opts *StoreOptions) (*Store, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}
	if opts.CmdContext == nil {
		return nil, fmt.Errorf("CmdContext function cannot be nil")
	}
	if opts.ClientOptions == nil {
		return nil, fmt.Errorf("ClientOptions function cannot be nil")
	}

	return &Store{opts: opts}, nil
}

func StartStore(ctx context.Context, opts *StoreOptions) (*Store, error) {
	s, err := NewStore(opts)
	if err != nil {
		return nil, err
	}

	if err := s.Start(ctx); err != nil {
		return nil, err
	}

	if err := s.PollReady(ctx); err != nil {
		s.GracefulShutdown(ctx)
		return nil, err
	}

	return s, nil
}
