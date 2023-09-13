package redisx

import (
	"context"
	"fmt"
	"time"
)

// delCmd delete命令
type delCmd struct {
	ctx context.Context
	key string

	err      error
	callback func(error)
}

func (c *delCmd) Do(cli *Client) error {
	c.err = cli.Del(c.key)
	return c.err
}

func (c *delCmd) Key() string {
	return c.key
}

func (c *delCmd) Context() context.Context {
	return c.ctx
}

type getCmd struct {
	ctx context.Context
	key string

	ret      *StringRet
	callback func(*StringRet)
}

func (c *getCmd) Do(cli *Client) error {
	c.ret = cli.Get(c.key)
	return c.ret.Err()
}

func (c *getCmd) Key() string {
	return c.key
}

func (c *getCmd) Callback() {
	if c.callback != nil {
		c.callback(c.ret)
	}
}

func (c *getCmd) Context() context.Context {
	return c.ctx
}

type setCmd struct {
	ctx   context.Context
	key   string
	value interface{}

	err error
}

func (c *setCmd) Do(cli *Client) error {
	c.err = cli.Set(c.key, c.value)
	return c.err
}

func (c *setCmd) Key() string {
	return c.key
}

func (c *setCmd) Context() context.Context {
	return c.ctx
}

type setNXReturnCmd struct {
	ctx   context.Context
	key   string
	value interface{}

	retOk    bool
	retV     string
	err      error
	callback func(ok bool, v string)
}

func (c *setNXReturnCmd) Context() context.Context {
	return c.ctx
}

func (c *setNXReturnCmd) Do(cli *Client) error {
	c.retOk, c.err = cli.SetNX(c.key, c.value)
	if c.err != nil {
		return c.err
	}
	// 如果设置ok了，返回设置的value
	if c.retOk {
		c.retV = fmt.Sprint(c.value)
		return nil
	}
	// 没设置成功则get出来
	c.retV = cli.Get(c.key).val
	return nil
}

func (c *setNXReturnCmd) Key() string {
	return c.key
}

func (c *setNXReturnCmd) Callback() {
	if c.callback != nil {
		c.callback(c.retOk, c.retV)
	}
}

type setNXEXReturnCmd struct {
	ctx    context.Context
	key    string
	value  interface{}
	expire time.Duration

	retOk    bool
	retV     string
	err      error
	callback func(ok bool, v string)
}

func (c *setNXEXReturnCmd) Context() context.Context {
	return c.ctx
}

func (c *setNXEXReturnCmd) Do(cli *Client) error {
	c.retOk, c.err = cli.SetNXEX(c.key, c.value, c.expire)
	if c.err != nil {
		return c.err
	}
	// 如果设置ok了，返回设置的value
	if c.retOk {
		c.retV = fmt.Sprint(c.value)
		return nil
	}
	// 没设置成功则get出来
	c.retV = cli.Get(c.key).val
	return nil
}

func (c *setNXEXReturnCmd) Key() string {
	return c.key
}

func (c *setNXEXReturnCmd) Callback() {
	if c.callback != nil {
		c.callback(c.retOk, c.retV)
	}
}

type setNXCmd struct {
	ctx   context.Context
	key   string
	value interface{}

	ret      bool
	err      error
	callback func(ret bool)
}

func (c *setNXCmd) Do(cli *Client) error {
	c.ret, c.err = cli.SetNX(c.key, c.value)
	return c.err
}

func (c *setNXCmd) Key() string {
	return c.key
}

func (c *setNXCmd) Callback() {
	if c.callback != nil {
		c.callback(c.ret)
	}
}

func (c *setNXCmd) Context() context.Context {
	return c.ctx
}

type hmsetCmd struct {
	ctx   context.Context
	key   string
	value MapReq

	err error
}

func (c *hmsetCmd) Do(cli *Client) error {
	c.err = cli.HMSet(c.key, c.value)
	return c.err
}

func (c *hmsetCmd) Key() string {
	return c.key
}

func (c *hmsetCmd) Context() context.Context {
	return c.ctx
}

type hmgetCmd struct {
	ctx   context.Context
	key   string
	field []string

	ret      *StringSliceRet
	callback func(ret *StringSliceRet)
}

func (c *hmgetCmd) Do(cli *Client) error {
	c.ret = cli.HMGet(c.key, c.field...)
	return c.ret.Err()
}

func (c *hmgetCmd) Key() string {
	return c.key
}

func (c *hmgetCmd) Callback() {
	if c.callback != nil {
		c.callback(c.ret)
	}
}

func (c *hmgetCmd) Context() context.Context {
	return c.ctx
}

type hdelCmd struct {
	ctx   context.Context
	key   string
	field []string

	err error
}

func (c *hdelCmd) Do(cli *Client) error {
	c.err = cli.HDel(c.key, c.field...)
	return c.err
}

func (c *hdelCmd) Key() string {
	return c.key
}

func (c *hdelCmd) Context() context.Context {
	return c.ctx
}

// doCmd server命令
type doCmd struct {
	ctx context.Context
	cmd []interface{}

	ret      *StringRet
	callback func(*StringRet)
}

func (c *doCmd) Context() context.Context {
	return c.ctx
}

func (c *doCmd) Do(cli *Client) error {
	ret, err := cli.cli.Do(context.Background(), c.cmd...)
	c.ret = &StringRet{
		err: err,
		val: ret,
	}
	return err
}

func (c *doCmd) Key() string {
	if len(c.cmd) > 0 {
		return fmt.Sprintf("%v", c.cmd[0])
	}
	return ""
}

func (c *doCmd) Callback() {
	if c.callback != nil {
		c.callback(c.ret)
	}
}
