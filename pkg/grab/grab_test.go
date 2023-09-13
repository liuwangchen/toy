package grab

import (
	"context"
	"fmt"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestGrab_Run(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"192.168.1.153:2379"},
	})
	if err != nil {
		t.Error(err)
		return
	}
	grab := NewGrab(client, "/testgrab/")
	b, err := grab.Run(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("grab data ", string(b))
	time.Sleep(time.Second * 5)
	grab.Close()
}
