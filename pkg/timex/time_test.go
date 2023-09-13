package timex

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetWeekDayInRangeTime(t *testing.T) {
	startTime := time.Unix(1650851940, 0)
	fmt.Println("start:", startTime, startTime.Weekday())
	endTime := time.Unix(1651334340, 0)
	fmt.Println("end:", endTime, endTime.Weekday())

	for i := 0; i < 6; i++ {
		times := GetWeekDayInRangeTime(startTime, endTime, time.Weekday(i))
		for _, v := range times {
			assert.Equal(t, time.Weekday(i), v.Weekday())
		}
	}
}
