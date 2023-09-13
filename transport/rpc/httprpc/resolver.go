package httprpc

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/liuwangchen/toy/pkg/endpoint"
	"github.com/liuwangchen/toy/registry"
	"github.com/liuwangchen/toy/selector"
)

// Target is resolver target
type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
	Path      string
}

func newTarget(endpoint string, insecure bool) (*Target, error) {
	if !strings.Contains(endpoint, "://") {
		if insecure {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	target := &Target{Scheme: u.Scheme, Authority: u.Host}
	if len(u.Path) > 1 {
		target.Endpoint = u.Path[1:]
		target.Path = u.Path
	}
	return target, nil
}

type resolver struct {
	selector.Selector
	registry.Watcher
	insecure bool
}

func newResolver(ctx context.Context, discovery registry.Discovery, target *Target, selector selector.Selector, insecure bool) (*resolver, error) {
	watcher, err := discovery.Watch(ctx, target.Endpoint)
	if err != nil {
		return nil, err
	}
	r := &resolver{
		Watcher:  watcher,
		Selector: selector,
		insecure: insecure,
	}
	done := make(chan error, 1)
	go func() {
		for {
			services, err := watcher.Next()
			if err != nil {
				done <- err
				return
			}
			if r.update(services) {
				done <- nil
				return
			}
		}
	}()
	select {
	case err := <-done:
		if err != nil {
			err := watcher.Stop()
			return nil, err
		}
	case <-ctx.Done():
		err := watcher.Stop()
		if err != nil {
			return nil, err
		}
		return nil, ctx.Err()
	}
	go func() {
		for {
			services, err := watcher.Next()
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				time.Sleep(time.Second)
				continue
			}
			r.update(services)
		}
	}()
	return r, nil
}

func (r *resolver) update(services []*registry.ServiceInstance) bool {
	nodes := make([]selector.Node, 0)
	for _, ins := range services {
		hostPort, err := endpoint.ExtractHostPortFromEndpoints(ins.Endpoints, "http", !r.insecure)
		if err != nil {
			continue
		}
		if hostPort == "" {
			continue
		}
		nodes = append(nodes, selector.NewNode("http", hostPort, ins))
	}
	if len(nodes) == 0 {
		return false
	}
	r.Apply(nodes)
	return true
}

func (r *resolver) Close() error {
	return r.Stop()
}
