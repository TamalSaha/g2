package runtime

import (
	"github.com/appscode/errors"
	"reflect"
	"testing"
)

func TestNewSchedTime(t *testing.T) {
	sampleTestCase := []struct {
		given    specScheduleTime
		expected specScheduleTime
		err      error
	}{
		{
			given:    specScheduleTime{"*", "2", "3", "4", "5", ""},
			expected: specScheduleTime{"*", "2", "3", "4", "5", "* 2 3 4 5"},
			err:      nil,
		},
		{
			given:    specScheduleTime{"", "6", "6", "2", "2", ""},
			expected: specScheduleTime{"*", "6", "6", "2", "2", "* 6 6 2 2"},
			err:      nil,
		},
		{
			given:    specScheduleTime{".", "2", "3", "4", "5", ""},
			expected: specScheduleTime{},
			err:      errors.NewGoError("invalid cron expression"),
		},
		{
			given:    specScheduleTime{"60", "2", "3", "4", "5", ""},
			expected: specScheduleTime{},
			err:      errors.NewGoError("invalid cron expression"),
		},
		{
			given:    specScheduleTime{"a", "2", "3", "4", "5", ""},
			expected: specScheduleTime{},
			err:      errors.NewGoError("invalid cron expression"),
		},
	}

	for _, tc := range sampleTestCase {
		got, err := NewSchedule(tc.given.Minute, tc.given.Hour, tc.given.Day, tc.given.Month, tc.given.WeekDay)
		if !reflect.DeepEqual(err, tc.err) {
			t.Fatalf("expected `%v` error but got `%v` error", tc.err, err)
		}
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("expected %+v but got %+v", tc.expected, got)
		}
	}
}

func TestNewScheduleFromExpression(t *testing.T) {
	sampleTestCase := []struct {
		given    string
		expected specScheduleTime
		err      error
	}{
		{
			given:    "* 2 3 4 5",
			expected: specScheduleTime{"*", "2", "3", "4", "5", "* 2 3 4 5"},
			err:      nil,
		},
		{
			given:    "2 3 4 5",
			expected: specScheduleTime{},
			err:      errors.NewGoError("invalid cron expression"),
		},
		{
			given:    "60 1 2 3 4",
			expected: specScheduleTime{},
			err:      errors.NewGoError("invalid cron expression"),
		},
		{
			given:    "a 2 3 5 6",
			expected: specScheduleTime{},
			err:      errors.NewGoError("invalid cron expression"),
		},
	}
	for _, tc := range sampleTestCase {
		got, err := NewScheduleFromExpression(tc.given)
		if !reflect.DeepEqual(err, tc.err) {
			t.Fatalf("expected `%v` error but got `%v` error", tc.err, err)
		}
		if !reflect.DeepEqual(tc.expected, got) {
			t.Fatalf("expected %+v but got %+v", tc.expected, got)
		}
	}
}

func TestSpecScheduleTimeBytes(t *testing.T) {
	exp := []byte{50, 53, 0, 50, 0, 50, 55, 0, 0, 0}
	sc, err := NewSchedule("25", "2", "27", "*", "*")
	if err != nil {
		t.Fatalf("no error expected but got `%v` error\n", err)
	}
	if !reflect.DeepEqual(exp, sc.Bytes()) {
		t.Fatalf("Expected %v, got %v\n", exp, sc.Bytes())
	}
}
