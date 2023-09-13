package etcdx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	EtcdAddr = "192.168.1.153:2379"
)

func TestClient(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{EtcdAddr},
		DialTimeout: time.Second * 5,
	})
	assert.Nil(t, err)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	defer cancelFunc()
	// 尝试连接 etcd 服务
	if _, err := cli.Status(ctx, cli.Endpoints()[0]); err != nil {
		t.Errorf("etcd client connection failed: %v", err)
	}
	cli.Close()
}
