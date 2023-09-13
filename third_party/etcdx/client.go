package etcdx

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/liuwangchen/toy/logger"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type KV struct {
	Key   string
	Value string
}

// Config 配置
type Config struct {
	Servers        string `xml:"servers" yaml:"servers"`
	DialTimeout    int64  `xml:"dial_timeout" yaml:"dial_timeout"`
	RequestTimeout int64  `xml:"request_timeout" yaml:"request_timeout"`
}

func defaultConfig(cfg *Config) {
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 10
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 10
	}
}

// keyConfig 单调键值的配置
type keyConfig struct {
	prefix bool // 前缀监听
	call   func(string, string, WatchEventType)
}

// Client etcd 客户端
type Client struct {
	once      sync.Once
	closeChan chan struct{}
	client    *clientv3.Client // etcd v3 client
	cfg       Config
	keys      *sync.Map
}

// NewClient 构造一个注册服务
func NewClient(client *clientv3.Client) *Client {
	cfg := Config{}
	defaultConfig(&cfg)
	c := &Client{
		closeChan: make(chan struct{}),
		client:    client,
		cfg:       cfg,
		keys:      new(sync.Map),
	}

	return c
}

func NewClientWithConfig(cfg Config) (*Client, error) {
	if len(cfg.Servers) == 0 {
		return nil, clientv3.ErrNoAvailableEndpoints
	}
	defaultConfig(&cfg)
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(cfg.Servers, ","),
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(cfg.DialTimeout)*time.Second)
	defer cancelFunc()
	// 尝试连接 etcd 服务
	_, err = client.Status(ctx, client.Endpoints()[0])
	if err != nil {
		return nil, err
	}

	cli := NewClient(client)
	cli.cfg = cfg
	return cli, nil
}

func (c *Client) GetPrefix(ctx context.Context, key string) (map[string][]byte, error) {
	resp, err := c.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	ret := make(map[string][]byte, len(resp.Kvs))
	for _, v := range resp.Kvs {
		ret[string(v.Key)] = v.Value
	}
	return ret, nil
}

// 设置kv
func (c *Client) Put(ctx context.Context, key string, value string) error {
	return c.PutWithTTL(ctx, key, value, 0)
}

// 批量设置kv
func (c *Client) PutBatch(ctx context.Context, kv map[string]string) error {
	ops := make([]clientv3.Op, 0, len(kv))
	for key, value := range kv {
		ops = append(ops, clientv3.OpPut(key, value))
	}
	_, err := c.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return err
	}
	return nil
}

// 批量设置kv，带lease的
func (c *Client) PutBatchWithLease(ctx context.Context, kv map[string]string, leaseId clientv3.LeaseID) error {
	ops := make([]clientv3.Op, 0, len(kv))
	for key, value := range kv {
		ops = append(ops, clientv3.OpPut(key, value, clientv3.WithLease(leaseId)))
	}
	_, err := c.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return err
	}
	return nil
}

// 设置kv，将value序列化成json格式存储
func (c *Client) PutJSON(ctx context.Context, key string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Put(ctx, key, string(b))
}

// PutWithTTL 带ttl的put
func (c *Client) PutWithTTL(ctx context.Context, key, value string, ttl int64) error {
	var opt []clientv3.OpOption
	if ttl > 0 {
		lease := clientv3.NewLease(c.client)
		grantResp, err := lease.Grant(ctx, ttl)
		if err != nil {
			return err
		}
		opt = append(opt, clientv3.WithLease(grantResp.ID))
	}

	_, err := c.client.Put(ctx, key, value, opt...)
	return err
}

// PutWithLease 带lease的put
func (c *Client) PutWithLease(ctx context.Context, key, value string, leaseId clientv3.LeaseID) error {
	_, err := c.client.Put(ctx, key, value, clientv3.WithLease(leaseId))
	return err
}

// PutWithKeepLive 设置一个带ttl的key，并且不断的keeplive使这个key不过期，当程序退出或者revoke这个lease这个key会删除
func (c *Client) PutWithKeepLive(ctx context.Context, key string, val string, ttl int64) (clientv3.LeaseID, error) {
	if ttl == 0 {
		return 0, errors.New("ttl can not set 0")
	}

	leaseResp, err := c.client.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}

	_, err = c.client.Put(ctx, key, val, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return 0, err
	}

	keepAliveChan, err := c.client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return 0, err
	}
	go func() {
		for range keepAliveChan {
		}
	}()
	return leaseResp.ID, nil
}

// CreateKeepLiveLease 创建带保活的租约
func (c *Client) CreateKeepLiveLease(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	if ttl == 0 {
		return 0, errors.New("ttl can not set 0")
	}

	leaseResp, err := c.client.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}

	keepAliveChan, err := c.client.KeepAlive(context.Background(), leaseResp.ID)
	if err != nil {
		return 0, err
	}
	go func() {
		for range keepAliveChan {
		}
	}()
	return leaseResp.ID, nil
}

// PutBatchWithKeepLive 批量设置一个带ttl的kvs，并且不断的keeplive使这一批key不过期，当程序退出或者revoke这个lease这批key会删除
func (c *Client) PutBatchWithKeepLive(ctx context.Context, kv map[string]string, ttl int64) (clientv3.LeaseID, error) {
	if ttl == 0 {
		return 0, errors.New("ttl can not set 0")
	}

	leaseResp, err := c.client.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}

	ops := make([]clientv3.Op, 0, len(kv))
	for key, value := range kv {
		// 批量设置带lease的
		ops = append(ops, clientv3.OpPut(key, value, clientv3.WithLease(leaseResp.ID)))
	}
	_, err = c.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return 0, err
	}

	// 保活
	keepAliveChan, err := c.client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return 0, err
	}
	go func() {
		for range keepAliveChan {
		}
	}()
	return leaseResp.ID, nil
}

func (c *Client) PutJSONWithTTL(ctx context.Context, key string, data interface{}, ttl int64) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.PutWithTTL(ctx, key, string(b), ttl)
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	ret, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(ret.Kvs) == 0 {
		return nil, nil
	}
	return ret.Kvs[0].Value, nil
}

func (c *Client) ExistKey(ctx context.Context, key string) (bool, error) {
	ret, err := c.client.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if len(ret.Kvs) == 0 {
		return false, nil
	}
	return true, nil
}

func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	ret, err := c.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

func (c *Client) GetKeys(ctx context.Context, key string) ([]string, error) {
	response, err := c.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(response.Kvs))
	for _, kv := range response.Kvs {
		res = append(res, string(kv.Key))
	}

	return res, nil
}

func (c *Client) GetLeaseTTl(ctx context.Context, id int) (int64, error) {
	response, err := c.client.TimeToLive(ctx, clientv3.LeaseID(id))
	if err != nil {
		return 0, err
	}
	return response.TTL, nil
}

func (c *Client) GetPrefixValues(ctx context.Context, key string) ([][]byte, error) {
	ret, err := c.GetPrefix(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, nil
	}
	tmp := make([][]byte, 0, len(ret))
	for _, v := range ret {
		tmp = append(tmp, v)
	}
	return tmp, nil
}

func (c *Client) GetJSON(ctx context.Context, key string, data interface{}) error {
	b, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, data)
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key)
	return err
}

func (c *Client) DeletePrefix(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, key, clientv3.WithPrefix())
	return err
}

// PutIfNotExist 如果create不存在,放入键值create跟kv
func (c *Client) PutIfNotExist(ctx context.Context, key, value string, opts ...clientv3.OpOption) (bool, error) {
	// 比较Revision, 当key不存在时，createRevision是0
	keyNotExist := clientv3.Compare(clientv3.CreateRevision(key), "=", 0)

	var puts = make([]clientv3.Op, 0, 1)
	puts = append(puts, clientv3.OpPut(key, value, opts...))

	resp, err := c.client.Txn(ctx).If(keyNotExist).Then(puts...).Commit()
	if err != nil {
		return false, err
	}
	if !resp.Succeeded {
		return false, nil
	}
	return true, nil
}

// UpdateTx 更新事务
func (c *Client) UpdateTx(ctx context.Context, key, val string) error {
	// 比较Revision, 当key不存在时，createRevision是0
	keyExist := clientv3.Compare(clientv3.CreateRevision(key), "!=", 0)
	put := clientv3.OpPut(key, val)
	resp, err := c.client.Txn(ctx).If(keyExist).Then(put).Commit()
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return errors.New("update failed")
	}
	return nil
}

// WatchEventType 事件类型
type WatchEventType int

const (
	Init   WatchEventType = iota // 初始
	Create                       // 创建
	Modify                       // 修改
	Delete                       // 删除
)

func (c *Client) AddWatch(key string, prefix bool, cb func(string, string, WatchEventType)) bool {
	return c.AddWatchWithContext(context.Background(), key, prefix, cb)
}

func (c *Client) AddWatchWithContext(ctx context.Context, key string, prefix bool, cb func(string, string, WatchEventType)) bool {
	if prefix {
		if key[len(key)-1] != '/' {
			key += "/"
		}
	}
	if _, ok := c.keys.Load(key); ok {
		return false
	}
	c.keys.Store(key, &keyConfig{
		prefix: prefix,
		call:   cb,
	})
	go c.addWatch(ctx, key, prefix, cb)
	return true
}

func (c *Client) addWatch(ctx context.Context, key string, prefix bool, cb func(string, string, WatchEventType)) {
	defer c.keys.Delete(key)
	reqctx, cancel := context.WithTimeout(ctx, time.Duration(c.cfg.RequestTimeout)*time.Second)
	var ops []clientv3.OpOption
	if prefix {
		ops = append(ops, clientv3.WithPrefix())
	}
	resp, err := c.client.Get(reqctx, key, ops...)
	if err != nil {
		panic(err)
	}
	cancel()
	for _, ev := range resp.Kvs {
		cb(string(ev.Key), string(ev.Value), Init)
	}
	var rch clientv3.WatchChan
	if prefix {
		rch = c.client.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithRev(resp.Header.Revision+1))
	} else {
		rch = c.client.Watch(ctx, key, clientv3.WithRev(resp.Header.Revision+1))
	}
	logger.Debug("[Client] AddWatch start watch [%s] prefix[%v]", key, prefix)
	for {
		select {
		case <-c.closeChan:
			return
		case <-c.client.Ctx().Done():
			return
		case <-ctx.Done():
			return
		case wresp := <-rch:
			err := wresp.Err()
			if err != nil {
				logger.Error("[Client] watch %s response error: %s ", key, err.Error())
				return
			}
			logger.Debug("[Client] watch %s response %+v", key, wresp)
			for _, ev := range wresp.Events {
				logger.Debug("[Client] watch wresp.Events key:%s value:%s", string(ev.Kv.Key), string(ev.Kv.Value))
				if ev.IsCreate() {
					cb(string(ev.Kv.Key), string(ev.Kv.Value), Create)
				} else if ev.IsModify() {
					cb(string(ev.Kv.Key), string(ev.Kv.Value), Modify)
				} else if ev.Type == mvccpb.DELETE {
					cb(string(ev.Kv.Key), string(ev.Kv.Value), Delete)
				}
			}
		}
	}
}

// Close 停止
func (c *Client) Close() {
	c.once.Do(func() {
		close(c.closeChan)
		c.client.Close()
	})
}

// 撤销lease
func (c *Client) RevokeLease(ctx context.Context, id clientv3.LeaseID) error {
	_, err := c.client.Revoke(ctx, id)
	return err
}
