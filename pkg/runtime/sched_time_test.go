package runtime

import (
	"testing"
	"reflect"
)

func TestNewSchedTime(t *testing.T) {
	ex := SpecScheduleTime{
		Minute: "1",
		Hour: "2",
		Day: "3",
		Month: "4",
		WeekDay: "5",
	}
	act := NewSchedTime(ex.Minute, ex.Hour, ex.Day, ex.Month, ex.WeekDay)
	if !reflect.DeepEqual(ex, act) {
		t.Errorf("Expected %v, got %v\n", ex, act)
	}
}

func TestSpecScheduleTimeBytes(t *testing.T) {
	exp := []byte{50, 53, 0, 50, 0, 50, 55, 0, 42, 0, 42, 0}
	sc := SpecScheduleTime{
		Minute: "25",
		Hour: "2",
		Day: "27",
		Month: "*",
		WeekDay: "*",
	}
	if !reflect.DeepEqual(exp, sc.Bytes()) {
		t.Errorf("Expected %v, got %v\n", exp, sc.Bytes())
	}
}