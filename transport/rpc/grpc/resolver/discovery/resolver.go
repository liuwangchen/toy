package discovery

import (
	"context"
	"errors"
	"time"

	"github.com/liuwangchen/toy/logger"
	"github.com/liuwangchen/toy/pkg/endpoint"
	"github.com/liuwangchen/toy/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

type discoveryResolver struct {
	w  registry.Watcher
	cc resolver.ClientConn

	ctx    context.Context
	cancel context.CancelFunc

	insecure         bool
	debugLogDisabled bool
}

func (r *discoveryResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}
		ins, err := r.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			logger.Error("[resolver] Failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}
		r.update(ins)
	}
}

func (r *discoveryResolver) update(ins []*registry.ServiceInstance) {
	addrs := make([]resolver.Address, 0)
	endpoints := make(map[string]struct{})
	for _, in := range ins {
		endpoint, err := endpoint.ExtractHostPortFromEndpoints(in.Endpoints, "grpc", !r.insecure)
		if err != nil {
			logger.Error("[resolver] Failed to parse discovery endpoint: %v", err)
			continue
		}
		if endpoint == "" {
			continue
		}
		// filter redundant endpoints
		if _, ok := endpoints[endpoint]; ok {
			continue
		}
		endpoints[endpoint] = struct{}{}
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata),
			Addr:       endpoint,
		}
		addr.Attributes = addr.Attributes.WithValue("rawServiceInstance", in)
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		logger.Warn("[resolver] Zero endpoint found,refused to write, instances: %v", ins)
		return
	}
	_ = r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *discoveryResolver) Close() {
	r.cancel()
	r.w.Stop()
}

func (r *discoveryResolver) ResolveNow(options resolver.ResolveNowOptions) {}

func parseAttributes(md map[string]string) *attributes.Attributes {
	var a *attributes.Attributes
	for k, v := range md {
		if a == nil {
			a = attributes.New(k, v)
		} else {
			a = a.WithValue(k, v)
		}
	}
	return a
}
