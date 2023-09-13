package redisx

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Server       string `json:"server"`
	Auth         string `json:"auth"`
	Index        int    `json:"index"`
	MinIdleConns int    `json:"min_idle_conns"`
	MaxConns     int    `json:"max_conns"`
	Tls          bool   `json:"tls"`
}

type Client struct {
	cli *goRedis
}

func NewClient(cfg Config) (*Client, error) {
	cli := newGoRedis(cfg)
	if err := cli.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("redis ping error: %v", err)
	}
	return &Client{
		cli: cli,
	}, nil
}

func (c *Client) Close() error {
	return c.cli.Close()
}

func (c *Client) Get(key string) *StringRet {
	return c.cli.Get(context.Background(), key)
}

func (c *Client) MGet(keys ...string) *StringSliceRet {
	if len(keys) == 0 {
		return &StringSliceRet{
			err: fmt.Errorf("keys is empty"),
		}
	}
	return c.cli.MGet(context.Background(), keys...)
}

func (c *Client) ZAdd(key string, score ZAddReq) (int64, error) {
	if score.Count() == 0 {
		return 0, fmt.Errorf("score is empty")
	}
	return c.cli.ZAdd(context.Background(), key, score)
}

func (c *Client) ZRange(key string, start, stop int64) *StringSliceRet {
	return c.cli.ZRange(context.Background(), key, start, stop)
}

// Set 设置值
// @param value - 如果为proto.Message, 则自动序列化
func (c *Client) Set(key string, value interface{}) error {
	data, err := Encode(value)
	if err != nil {
		return err
	}
	return c.cli.Set(context.Background(), key, data, 0)
}

func (c *Client) SetEX(key string, value interface{}, expire time.Duration) error {
	data, err := Encode(value)
	if err != nil {
		return err
	}
	return c.cli.Set(context.Background(), key, data, expire)
}

func (c *Client) SetNX(key string, value interface{}) (bool, error) {
	data, err := Encode(value)
	if err != nil {
		return false, err
	}
	return c.cli.SetNX(context.Background(), key, data, 0)
}

func (c *Client) SetNXEX(key string, value interface{}, expire time.Duration) (bool, error) {
	data, err := Encode(value)
	if err != nil {
		return false, err
	}
	return c.cli.SetNX(context.Background(), key, data, expire)
}

func (c *Client) HGetAll(key string) *MapRet {
	return c.cli.HGetAll(context.Background(), key)
}

func (c *Client) HMSet(key string, value MapReq) error {
	return c.cli.HMSet(context.Background(), key, value)
}

func (c *Client) HMGet(key string, field ...string) *StringSliceRet {
	return c.cli.HMGet(context.Background(), key, field...)
}

func (c *Client) HDel(key string, fields ...string) error {
	return c.cli.HDel(context.Background(), key, fields...)
}

func (c *Client) Del(keys ...string) error {
	return c.cli.Del(context.Background(), keys...)
}

func (c *Client) HVals(key string) *StringSliceRet {
	return c.cli.HVals(context.Background(), key)
}

// goRedis 接入 github.com/go-redis/redis/v8
type goRedis struct {
	rdb *redis.Client
}

func newGoRedis(cfg Config) *goRedis {
	opt := &redis.Options{
		Addr:         cfg.Server,
		Password:     cfg.Auth,
		DB:           cfg.Index,
		MaxRetries:   2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  -1 * time.Second, // 不超时
		WriteTimeout: -1 * time.Second, // 不超时
		MinIdleConns: cfg.MinIdleConns,
		PoolSize:     cfg.MaxConns,
	}
	if cfg.Tls {
		opt.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	rdb := redis.NewClient(opt)

	return &goRedis{
		rdb: rdb,
	}
}

func (c *goRedis) parseError(err error) error {
	if err == redis.Nil {
		return nil
	}
	return err
}

func (c *goRedis) sliceCmdConvert(cmd *redis.SliceCmd) *StringSliceRet {
	var vals = make([]string, 0, len(cmd.Val()))
	for _, v := range cmd.Val() {
		switch val := v.(type) {
		case string:
			vals = append(vals, val)
		case nil:
			vals = append(vals, "")
		default:
			return &StringSliceRet{
				err: fmt.Errorf("not support type: %T", val),
			}
		}
	}
	return &StringSliceRet{
		err: c.parseError(cmd.Err()),
		val: vals,
	}
}

func (c *goRedis) Do(ctx context.Context, args ...interface{}) (string, error) {
	ret := c.rdb.Do(ctx, args...)
	if ret.Err() != nil {
		return "", ret.Err()
	}
	return ret.String(), nil
}

func (c *goRedis) Ping(ctx context.Context) error {
	cmd := c.rdb.Ping(ctx)
	if err := cmd.Err(); err != nil {
		return err
	}
	return nil
}

func (c *goRedis) Close() error {
	return c.rdb.Close()
}

func (c *goRedis) Get(ctx context.Context, key string) *StringRet {
	val := c.rdb.Get(ctx, key)
	return &StringRet{
		err: c.parseError(val.Err()),
		val: val.Val(),
	}
}

func (c *goRedis) MGet(ctx context.Context, keys ...string) *StringSliceRet {
	cmd := c.rdb.MGet(ctx, keys...)
	return c.sliceCmdConvert(cmd)
}

// Set 设置值
// @param value - 如果为proto.Message, 则自动序列化
func (c *goRedis) Set(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	cmd := c.rdb.Set(ctx, key, value, expire)
	return c.parseError(cmd.Err())
}

func (c *goRedis) SetNX(ctx context.Context, key string, value interface{}, expire time.Duration) (bool, error) {
	cmd := c.rdb.SetNX(ctx, key, value, expire)
	return cmd.Val(), c.parseError(cmd.Err())
}

func (c *goRedis) HGetAll(ctx context.Context, key string) *MapRet {
	cmd := c.rdb.HGetAll(ctx, key)
	return &MapRet{
		err: c.parseError(cmd.Err()),
		val: cmd.Val(),
	}
}

func (c *goRedis) HMSet(ctx context.Context, key string, value MapReq) error {
	cmd := c.rdb.HMSet(ctx, key, value...)
	return c.parseError(cmd.Err())
}

func (c *goRedis) HMGet(ctx context.Context, key string, field ...string) *StringSliceRet {
	cmd := c.rdb.HMGet(ctx, key, field...)
	return c.sliceCmdConvert(cmd)
}

func (c *goRedis) HDel(ctx context.Context, key string, fields ...string) error {
	cmd := c.rdb.HDel(ctx, key, fields...)
	return c.parseError(cmd.Err())
}

func (c *goRedis) Del(ctx context.Context, keys ...string) error {
	cmd := c.rdb.Del(ctx, keys...)
	return c.parseError(cmd.Err())
}

func (c *goRedis) HVals(ctx context.Context, key string) *StringSliceRet {
	cmd := c.rdb.HVals(ctx, key)
	return &StringSliceRet{
		err: c.parseError(cmd.Err()),
		val: cmd.Val(),
	}
}

func (c *goRedis) ZAdd(ctx context.Context, key string, data map[float64]interface{}) (int64, error) {
	var memebers = make([]*redis.Z, 0, len(data))
	for score, member := range data {
		memebers = append(memebers, &redis.Z{
			Score:  score,
			Member: member,
		})
	}
	cmd := c.rdb.ZAdd(ctx, key, memebers...)
	return cmd.Val(), c.parseError(cmd.Err())
}

func (c *goRedis) ZRange(ctx context.Context, key string, start, stop int64) *StringSliceRet {
	cmd := c.rdb.ZRange(ctx, key, start, stop)
	return &StringSliceRet{
		err: c.parseError(cmd.Err()),
		val: cmd.Val(),
	}
}
