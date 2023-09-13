package executor

import (
	"context"
	"errors"
	"os"
	"os/signal"
)

type options struct {
	args    []interface{}
	before  Executor
	after   Executor
	signals map[os.Signal]Executor
}

type Opt func(*options)

func WithArguments(args ...interface{}) Opt {
	return func(opts *options) {
		opts.args = args
	}
}

func WithSignal(exec Executor, sigs ...os.Signal) Opt {
	return func(opts *options) {
		for _, sig := range sigs {
			opts.signals[sig] = exec
		}
	}
}

func WithBefore(exec Executor) Opt {
	return func(opts *options) {
		opts.before = exec
	}
}

func WithAfter(exec Executor) Opt {
	return func(opts *options) {
		opts.after = exec
	}
}

type Exec struct {
	opts *options
	exec Executor
}

func New(exec Executor, opts ...Opt) *Exec {
	mopts := &options{
		args:    []interface{}{},
		signals: map[os.Signal]Executor{},
	}
	for _, opt := range opts {
		opt(mopts)
	}
	routine := &Exec{
		opts: mopts,
		exec: exec,
	}
	return routine
}

func (r *Exec) Do(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context required")
	}

	// before
	err := r.before(ctx)
	if err != nil {
		return err
	}

	// after
	defer r.after()

	// arg
	if len(r.opts.args) > 0 {
		ctx = WithArgments(ctx, r.opts.args...)
	}

	// signal
	sigChan := r.sig()

	// do
	ch := r.do(ctx, r.exec)

	for {
		select {
		case err := <-ch:
			if err != nil {
				return err
			}
			return nil
		case sig := <-sigChan:
			executor, ok := r.opts.signals[sig]
			if !ok {
				break
			}
			err := executor.Execute(ctx)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (r *Exec) sig() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	if len(r.opts.signals) == 0 {
		return sigChan
	}
	s := make([]os.Signal, 0, len(r.opts.signals))
	for sig := range r.opts.signals {
		s = append(s, sig)
	}
	signal.Notify(sigChan, s...)
	return sigChan
}

func (r *Exec) before(ctx context.Context) error {
	if r.opts.before != nil {
		if err := r.opts.before.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *Exec) after() error {
	if r.opts.after != nil {
		if err := r.opts.after.Execute(context.Background()); err != nil {
			return err
		}
	}
	return nil
}

func (r *Exec) do(ctx context.Context, exec Executor) chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- exec.Execute(ctx)
	}()
	return ch
}

func Execute(ctx context.Context, exec Executor, opts ...Opt) error {
	return New(exec, opts...).Do(ctx)
}
