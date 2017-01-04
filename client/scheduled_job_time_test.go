package client

import (
	"reflect"
	"testing"

	rt "github.com/appscode/g2/pkg/runtime"
)

func TestSchedTimeWithDataBytes(t *testing.T) {
	original := []byte{50, 53, 0, 50, 0, 50, 55, 0, 42, 0, 42, 0, 72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}
	scd := SchedTimeWithData{
		SpecScheduleTime: rt.SpecScheduleTime{
			Minute:  "25",
			Hour:    "2",
			Day:     "27",
			Month:   "*",
			WeekDay: "*",
		},
		data: TestStr,
	}
	if !reflect.DeepEqual(original, scd.Bytes()) {
		t.Errorf("Expected %v, got %v\n", original, scd.Bytes())
	}
}
