package client

import (
	"reflect"
	"testing"

	rt "github.com/appscode/g2/pkg/runtime"
)

func TestSchedTimeWithDataBytes(t *testing.T) {
	expected := []byte{50, 53, 0, 50, 0, 50, 55, 0, 0, 0, 72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100}
	td, err := rt.NewScheduleFromExpression("25 2 27 * *")
	if err != nil {
		t.Fatal(err)
	}

	scd := ScheduleTimedData{
		scheduledTimeData: td.Bytes(),
		data:              []byte(TestStr),
	}
	got := scd.Bytes()
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("Expected %v, got %v\n", expected, got)
	}
}
