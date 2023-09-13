package grab

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/liuwangchen/toy/pkg/election"
	"github.com/liuwangchen/toy/pkg/ipx"
	"github.com/liuwangchen/toy/third_party/etcdx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Grab 抢饼干
// 分布式抢夺饼干，用于服务器无状态启动且能获取有状态数据的场景
type Grab struct {
	etcdxClient        *etcdx.Client
	dataPath           string
	el                 *election.Election
	currentGrabDataKey string
	keepLiveLeaseId    clientv3.LeaseID // 保证所有grab的key都用一个lease进行保活
	isClose            bool

	// option
	id                string
	onGrabDataKeyLose func(dataKey string)
	timeout           time.Duration
	ttl               int // 两个用途，一个是选举使用的ttl，一个是续租饼干的ttl
	retryCount        int
	retryDuration     time.Duration
}

func (this *Grab) setTTl(ttl int) {
	this.ttl = ttl
}

func (this *Grab) setTimeout(timeout time.Duration) {
	this.timeout = timeout
}

func (this *Grab) setOnGrabDataKeyLose(cb func(dataKey string)) {
	this.onGrabDataKeyLose = cb
}

func (this *Grab) setId(id string) {
	this.id = id
}

type isetOp interface {
	setId(id string)
	setOnGrabDataKeyLose(cb func(dataKey string))
	setTimeout(timeout time.Duration)
	setTTl(ttl int)
}

type option func(s isetOp)

func WithTTl(ttl int) option {
	return func(s isetOp) {
		s.setTTl(ttl)
	}
}

func WithId(id string) option {
	return func(s isetOp) {
		s.setId(id)
	}
}

func WithOnGrabDataKeyLose(cb func(dataKey string)) option {
	return func(s isetOp) {
		s.setOnGrabDataKeyLose(cb)
	}
}

func WithTimeout(timeout time.Duration) option {
	return func(s isetOp) {
		s.setTimeout(timeout)
	}
}

func NewGrab(client *clientv3.Client, dataPath string, opts ...option) *Grab {
	g := &Grab{
		etcdxClient:       etcdx.NewClient(client),
		dataPath:          dataPath,
		onGrabDataKeyLose: func(dataKey string) {},
		timeout:           time.Second * 30,
		ttl:               20,
		retryCount:        20,
		retryDuration:     time.Second,
	}
	for _, opt := range opts {
		opt(g)
	}
	if len(g.id) == 0 {
		// 用ip+进程id拼接
		g.id = fmt.Sprintf("%s_%d", ipx.GetOutboundIP(), os.Getpid())
	}

	g.el = election.New(client, dataPath, election.WithId(g.id), election.WithTTl(g.ttl))
	return g
}

func (this *Grab) Run(ctx context.Context) ([]byte, error) {
	var grabData []byte
	// 选举
	err := this.el.Run(ctx, func(ctx context.Context) {
		// 成为leader开始抢饼干，直到抢到饼干
		for {
			// 找到一个可以抢的饼干
			key, b, ok := this.findCanGrabKey()
			if !ok {
				// 找不到睡眠1秒
				time.Sleep(this.retryDuration)
				continue
			}

			// 设置拥有者
			ok, err := this.setOwner(key)
			if !ok || err != nil {
				// 找不到睡眠1秒
				time.Sleep(time.Second)
				continue
			}
			grabData = b
			break
		}
	})
	if err != nil {
		return nil, err
	}
	this.watchLose(ctx)
	return grabData, nil
}

// Close 退出时候释放抢到的饼干
func (this *Grab) Close() {
	this.isClose = true
	this.releaseAllGrab()
}

// 释放所有抢到的饼干
func (this *Grab) releaseAllGrab() {
	if this.keepLiveLeaseId > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), this.timeout)
		defer cancel()

		_ = this.etcdxClient.RevokeLease(ctx, this.keepLiveLeaseId)
	}
}

func (this *Grab) isGrab() bool {
	return len(this.currentGrabDataKey) > 0
}

// 是否饼干被抢到
func (this *Grab) hasGrab(dataKey string) bool {
	if this.currentGrabDataKey == dataKey {
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), this.timeout)
	defer cancel()

	exist, err := this.etcdxClient.ExistKey(ctx, this.grabPath(dataKey))
	if err != nil {
		return false
	}
	return exist
}

func (this *Grab) grabPath(dataKey string) string {
	return path.Join("/grab/", dataKey)
}

// 给饼干设置拥有者
func (this *Grab) setOwner(dataKey string) (bool, error) {
	grabPath := this.grabPath(dataKey)

	ctx, cancel := context.WithTimeout(context.Background(), this.timeout)
	defer cancel()

	leaseId, err := this.etcdxClient.CreateKeepLiveLease(ctx, int64(this.ttl))
	if err != nil {
		return false, err
	}

	ok, err := this.etcdxClient.PutIfNotExist(ctx, grabPath, this.id, clientv3.WithLease(leaseId))
	if err != nil {
		_ = this.etcdxClient.RevokeLease(ctx, leaseId)
		return false, err
	}

	if !ok {
		_ = this.etcdxClient.RevokeLease(ctx, leaseId)
		return false, nil
	}

	this.currentGrabDataKey = dataKey
	this.keepLiveLeaseId = leaseId
	return true, nil
}

// 获取可以抢的饼干
func (this *Grab) findCanGrabKey() (string, []byte, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), this.timeout)
	defer cancel()

	// 所有饼干key
	keys, err := this.etcdxClient.GetKeys(ctx, this.dataPath)
	if err != nil {
		return "", nil, false
	}

	// 已经抢到的饼干key
	grabKeys, err := this.etcdxClient.GetKeys(ctx, this.grabPath(this.dataPath))
	if err != nil {
		return "", nil, false
	}

	canGrabKey := ""
	for _, key := range keys {
		hasGrab := false
		for _, grabKey := range grabKeys {
			if strings.Contains(grabKey, key) {
				hasGrab = true
				break
			}
		}
		if !hasGrab {
			canGrabKey = key
			break
		}
	}
	if len(canGrabKey) == 0 {
		return "", nil, false
	}

	b, err := this.etcdxClient.Get(ctx, canGrabKey)
	if err != nil {
		return "", nil, false
	}
	return canGrabKey, b, true
}

// 监听饼干数据变化
func (this *Grab) watchLose(ctx context.Context) {
	grabPath := this.grabPath(this.currentGrabDataKey)
	this.etcdxClient.AddWatchWithContext(ctx, grabPath, false, func(key string, value string, eventType etcdx.WatchEventType) {
		if this.isClose {
			return
		}
		switch eventType {
		case etcdx.Delete:
			// 重试抢饼干
			for i := 0; i < this.retryCount; i++ {
				ok, err := this.setOwner(this.currentGrabDataKey)
				if err != nil {
					time.Sleep(this.retryDuration)
					continue
				}
				if !ok {
					// 没有抢到
					this.onGrabDataKeyLose(this.currentGrabDataKey)
				}
				break
			}
		}
	})
}
