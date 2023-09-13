package timer

import (
	"fmt"
	"time"

	"github.com/gorhill/cronexpr"
	"github.com/liuwangchen/toy/pkg/priority_queue"
)

var (
	Exp_EverySecond = "*/1 * * * * * *"
)

type TimerArg struct {
	TimerId uint32
	DoTime  time.Time
	Cancel  func()
}

type timerEntity struct {
	id       uint32
	doTime   time.Time           // 执行时间
	exp      string              // cron表达式
	f        func(arg *TimerArg) // 执行函数
	canceled bool
}

func (t *timerEntity) Key() string {
	return fmt.Sprint(t.id)
}

type ITimer interface {
	AddTimeFunc(t time.Time, f func(arg *TimerArg)) uint32
	AddExpFunc(exp string, f func(arg *TimerArg)) uint32
	RemoveTimer(id uint32)
	Now() time.Time
	Tick(now time.Time)
}

type timer struct {
	id uint32
	tq *priority_queue.PriorityQueue

	// option
	limit int
	now   time.Time
}

func (this *timer) setNow(now time.Time) {
	this.now = now
}

func (this *timer) setLimit(limit int) {
	this.limit = limit
}

type isetOp interface {
	setNow(now time.Time)
	setLimit(limit int)
}

type option func(s isetOp)

func WithNow(now time.Time) option {
	return func(s isetOp) {
		s.setNow(now)
	}
}

func WithLimit(limit int) option {
	return func(s isetOp) {
		s.setLimit(limit)
	}
}

// NewTimer 初始化manager，配置信息
func NewTimer(ops ...option) ITimer {
	t := &timer{
		now: time.Now(),
		tq: priority_queue.NewPriorityQueue(func(item1, item2 interface{}) bool {
			ex1 := item1.(*timerEntity)
			ex2 := item2.(*timerEntity)

			if ex1.doTime.Equal(ex2.doTime) {
				return ex1.id < ex2.id
			}

			return ex1.doTime.Before(ex2.doTime)
		}),
		limit: 1024,
	}
	for _, op := range ops {
		op(t)
	}
	return t
}

// AddExpFunc 添加执行任务，用crontab表达式的方式
func (this *timer) AddExpFunc(exp string, f func(arg *TimerArg)) uint32 {
	nextDoTime := cronexpr.MustParse(exp).Next(this.now)
	if nextDoTime.IsZero() {
		return 0
	}
	timer := &timerEntity{
		doTime: nextDoTime,
		exp:    exp,
		f:      f,
	}
	this.id++
	timer.id = this.id
	this.tq.Push(timer)
	return timer.id
}

// AddTimeFunc 添加执行任务，用具体时间
func (this *timer) AddTimeFunc(t time.Time, f func(arg *TimerArg)) uint32 {
	exp := fmt.Sprintf("%d %d %d %d %d * %d", t.Second(), t.Minute(), t.Hour(), t.Day(), t.Month(), t.Year())
	return this.AddExpFunc(exp, f)
}

func (this *timer) RemoveTimer(id uint32) {
	this.tq.RemoveByKey(fmt.Sprint(id))
}

// Tick 运行
func (this *timer) Tick(now time.Time) {
	// 更新manager的时间
	this.now = now
	dealCount := 0

	for this.tq.Len() > 0 {
		// 执行次数
		if dealCount > this.limit {
			break
		}

		// 如果堆顶执行时间大于当前时间，break
		if this.tq.Top().(*timerEntity).doTime.After(now) {
			break
		}

		t := this.tq.Pop().(*timerEntity)
		t.f(&TimerArg{
			TimerId: t.id,
			DoTime:  t.doTime,
			Cancel: func() {
				t.canceled = true
			},
		})
		dealCount++

		// 任务被取消了
		if t.canceled {
			continue
		}

		nextDoTime := cronexpr.MustParse(t.exp).Next(now)
		// 有可能不存在了
		if !nextDoTime.IsZero() {
			t.doTime = nextDoTime
			this.tq.Push(t)
		}
	}
}

func (this *timer) Now() time.Time {
	return this.now
}
