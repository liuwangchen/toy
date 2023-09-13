package async

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var fcMap map[string]chan func()

func init() {
	fcMap = make(map[string]chan func())
	for i := 0; i < 10; i++ {
		id := fmt.Sprint(i)
		fc := make(chan func())
		fcMap[id] = fc
		go func() {
			for f := range fc {
				f()
				fmt.Println(id, "do")
			}
		}()
	}
}

func pushChan(ctx context.Context, f func()) error {
	val := ctx.Value("id")
	if val == nil {
		return fmt.Errorf("id is empty")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case fcMap[val.(string)] <- f:
		return nil
	}
}

func TestMultiAsync(t *testing.T) {
	ma := New(pushChan)

	_, err := ma.Do(context.Background(), nil)
	assert.NotNil(t, err)

	ctx1 := context.WithValue(context.Background(), "id", "1")
	ctx2, cancel := context.WithTimeout(ctx1, time.Second)
	defer cancel()
	r, err := ma.Do(ctx2, func() (interface{}, error) {
		return 1, nil
	})
	assert.Nil(t, err)
	assert.EqualValues(t, 1, r)
}
