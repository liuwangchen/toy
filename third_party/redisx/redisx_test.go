package redisx

import (
	"context"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/liuwangchen/toy/pkg/async"
)

var (
	cfg = Config{
		Server: "127.0.0.1:6379",
		Auth:   "",
		Index:  1,
	}
	cli     *Client
	chanCli *AsyncClient
	funcCli *AsyncClient
)

func TestMain(m *testing.M) {
	var err error
	cli, err = NewClient(cfg)
	if err != nil {
		panic(err)
	}
	as := async.New(func(ctx context.Context, f func()) error {
		f()
		return nil
	})

	chanCli, err = NewAsyncWithConfig(cfg, as, WithMaxCmdQueue(1024))
	if err != nil {
		panic(err)
	}
	funcCli, _ = NewAsyncWithConfig(cfg, as, WithMaxCmdQueue(1024))

	os.Exit(m.Run())
}

func TestGetEmpty(t *testing.T) {
	s, err := cli.Get("keyempty").String()
	if err != nil {
		t.Error(err)
	}
	if s != "" {
		t.Error("expect empty string")
	}
}

func TestGet(t *testing.T) {
	if err := cli.Set("key", "value"); err != nil {
		t.Error(err)
	}

	s, err := cli.Get("key").String()
	if err != nil {
		t.Error(err)
	}
	if s != "value" {
		t.Error("expect empty string")
	}

	if err := cli.Set("keyint", 1); err != nil {
		t.Error(err)
	}

	if v, err := cli.Get("keyint").Int32(); err != nil || v != 1 {
		t.Error(err)
	}
}

func TestMGet(t *testing.T) {
	if err := cli.Set("mget:key1", "value1"); err != nil {
		t.Error(err)
	}
	if err := cli.Set("mget:key2", "value2"); err != nil {
		t.Error(err)
	}
	defer cli.Del("mget:key1")
	defer cli.Del("mget:key2")
	r := cli.MGet("mget:key1", "mget:key2")
	sr, err := r.Strings()
	if err != nil {
		t.Error(err)
	}
	if sr[0] != "value1" || sr[1] != "value2" {
		t.Error("expect value1")
	}
}

func TestClient_ZAdd(t *testing.T) {
	req := NewZAddReq()
	req.Add(1, "key1")
	req.Add(2, 10)
	if _, err := cli.ZAdd("zadd:key", req); err != nil {
		t.Error(err)
	}
	defer cli.Del("zadd:key")
}

func TestGetProto(t *testing.T) {
	old := &Uint32Wrapper{
		U32: 1567,
	}
	if err := cli.Set("keyproto", old); err != nil {
		t.Error(err)
	}

	var s = &Uint32Wrapper{}
	err := cli.Get("keyproto").Proto(s)
	if err != nil {
		t.Error(err)
	}
	if s.U32 != old.U32 {
		t.Error("not equal")
	}
}

type Uint32Wrapper struct {
	U32                  uint32   `protobuf:"varint,1,opt,name=u32,proto3" json:"u32,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Uint32Wrapper) Reset()         { *m = Uint32Wrapper{} }
func (m *Uint32Wrapper) String() string { return proto.CompactTextString(m) }
func (*Uint32Wrapper) ProtoMessage()    {}
func (*Uint32Wrapper) Descriptor() ([]byte, []int) {
	return nil, []int{0}
}

func (m *Uint32Wrapper) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Uint32Wrapper.Unmarshal(m, b)
}
func (m *Uint32Wrapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Uint32Wrapper.Marshal(b, m, deterministic)
}
func (m *Uint32Wrapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Uint32Wrapper.Merge(m, src)
}
func (m *Uint32Wrapper) XXX_Size() int {
	return xxx_messageInfo_Uint32Wrapper.Size(m)
}
func (m *Uint32Wrapper) XXX_DiscardUnknown() {
	xxx_messageInfo_Uint32Wrapper.DiscardUnknown(m)
}

var xxx_messageInfo_Uint32Wrapper proto.InternalMessageInfo

func (m *Uint32Wrapper) GetU32() uint32 {
	if m != nil {
		return m.U32
	}
	return 0
}

func TestHGetallProto(t *testing.T) {
	var oldH = make(map[uint64]*Uint32Wrapper)
	oldH[1] = &Uint32Wrapper{
		U32: 1567,
	}

	var values = NewMapReq(4)
	for k, v := range oldH {
		values.Add(k, v)
	}

	if err := cli.HMSet("hkey", values); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := cli.Del("hkey"); err != nil {
			t.Error(err)
		}
	}()

	var h = make(map[uint64]*Uint32Wrapper)
	err := cli.HGetAll("hkey").ForEachUint64Bytes(func(k uint64, v []byte) error {
		r := &Uint32Wrapper{}
		if err := proto.Unmarshal(v, r); err != nil {
			return err
		}
		h[k] = r
		return nil
	})
	if err != nil || h[1].U32 != 1567 {
		t.Error(err)
	}
}

func TestHVals(t *testing.T) {
	var h = make(map[string]string)
	h["field"] = "value"
	cli.HMSet("keyvals", NewMapReqStringString(h))
	s1, _ := cli.HVals("keyvals").Strings()
	_ = s1

	var s []*Uint32Wrapper
	cli.HVals("keyvals").ForEachBytes(func(v []byte) error {
		m := &Uint32Wrapper{}
		if err := proto.Unmarshal(v, m); err != nil {
			return err
		}
		s = append(s, m)
		return nil
	})
}
