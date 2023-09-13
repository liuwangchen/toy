package timer

import (
	"fmt"
	"testing"
	"time"
)

func TestTimerManager_Run(t *testing.T) {
	tm := NewTimer()
	count := 0
	tm.AddExpFunc(Exp_EverySecond, func(arg *TimerArg) {
		fmt.Println(arg)
		count++
		if count == 10 {
			arg.Cancel()
		}
	})
	ticker := time.NewTicker(time.Millisecond * 100)
	for t := range ticker.C {
		tm.Tick(t)
	}
}
