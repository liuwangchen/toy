// Package election 选举胜利者获得执行权
package election

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/liuwangchen/toy/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// Election 选举器
// 优胜者拥有执行的作业的权利
type Election struct {
	client       *clientv3.Client
	electionPath string
	// SessionTTL 心跳时间，默认为10s
	ttl int
	id  string
}

func (this *Election) setId(id string) {
	this.id = id
}

func (this *Election) setTTl(ttl int) {
	this.ttl = ttl
}

type isetOp interface {
	setId(id string)
	setTTl(ttl int)
}

type option func(s isetOp)

func WithId(id string) option {
	return func(s isetOp) {
		s.setId(id)
	}
}

func WithTTl(ttl int) option {
	return func(s isetOp) {
		s.setTTl(ttl)
	}
}

// New 构造器
// electionPath 选举路径，用于区分不同的选举。用作etcd key例如: /election/{electionPath}
func New(client *clientv3.Client, electionPath string, opts ...option) *Election {
	e := &Election{
		electionPath: path.Join("/election/", electionPath),
		client:       client,
		ttl:          10,
	}
	for _, opt := range opts {
		opt(e)
	}
	if len(e.id) == 0 {
		// 用ip+进程id拼接
		hostname, _ := os.Hostname()
		e.id = fmt.Sprintf("%s_%d", hostname, os.Getpid())
	}

	return e
}

// Run 开始选举直到ctx被取消获得etcd的ctx被取消
// id 用于区分不同的选举者，会存在etcd value里。填相同的也不会有问题。空的话会使用hostname_pid
// fn 选举胜利者执行的作业
// 阻塞执行该函数，所有业务都要在该函数结束前停止
// 如果ctx被取消或者done被触发，必须立即退出作业
// 退出之后才会发生新的选举
func (this *Election) Run(ctx context.Context, fn func(ctx context.Context)) error {
	session, err := concurrency.NewSession(this.client, concurrency.WithTTL(this.ttl))
	if err != nil {
		return err
	}
	defer session.Close()

	el := concurrency.NewElection(session, this.electionPath)

	// 选举，非leader会卡住
	err = el.Campaign(ctx, this.id)
	if err != nil {
		return err
	}

	// leader执行任务
	logger.Info("[Election] %session:%session is leader", this.electionPath, this.id)
	fn(ctx)
	logger.Info("[Election] %session:%session func end", this.electionPath, this.id)
	return nil
}
