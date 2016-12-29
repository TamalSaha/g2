package server

import (
	"bytes"
	"testing"

	. "github.com/appscode/g2/pkg/runtime"
	"reflect"
)

func TestDecodeArgs(t *testing.T) {
	/*
		00 52 45 51                \0REQ        (Magic)
		00 00 00 07                7            (Packet type: SUBMIT_JOB)
		00 00 00 0d                13           (Packet length)
		72 65 76 65 72 73 65 00    reverse\0    (Function)
		00                         \0           (Unique ID)
		74 65 73 74                test         (Workload)
	*/

	data := []byte{
		0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65, 0x00,
		0x00,
		0x74, 0x65, 0x73, 0x74}
	slice, ok := decodeArgs(PT_SubmitJob, data)
	if !ok {
		t.Error("should be true")
	}

	if len(slice) != 3 {
		t.Error("arg count not match")
	}

	if !bytes.Equal(slice[0], []byte{0x72, 0x65, 0x76, 0x65, 0x72, 0x73, 0x65}) {
		t.Errorf("decode not match %+v", slice)
	}

	if !bytes.Equal(slice[1], []byte{}) {
		t.Error("decode not match")
	}

	if !bytes.Equal(slice[2], []byte{0x74, 0x65, 0x73, 0x74}) {
		t.Error("decode not match")
	}

	data = []byte{
		0x48, 0x3A, 0x2D, 0x51, 0x3A, 0x2D, 0x34, 0x37, 0x31, 0x36, 0x2D, 0x31,
		0x33, 0x39, 0x38, 0x31, 0x30, 0x36, 0x32, 0x33, 0x30, 0x2D, 0x32, 0x00, 0x00, 0x39, 0x38, 0x31,
		0x30,
	}

	slice, ok = decodeArgs(PT_WorkComplete, data)
	if !ok {
		t.Error("should be true")
	}

	if len(slice[0]) == 0 || len(slice[1]) == 0 {
		t.Error("arg count not match")
	}
}

func TestGetScheduleJobId(t *testing.T) {
	jobId := "H:-icee:-3043-1482985931-2"
	expectedSchedjobId := "S:-icee:-3043-1482985931-2"
	schedJobId := getScheduleJobId(jobId)
	if getScheduleJobId(jobId) != expectedSchedjobId {
		t.Errorf("Expected %s, got %s\n", expectedSchedjobId, schedJobId)
	}
}

func TestToSpecScheduleTime(t *testing.T) {
	args := &Tuple{
		t3: []byte{51, 49}, //Minute
		t4: []byte{50},     //Hour
		t5: []byte{52},     //Day Of Month
		t6: []byte{49, 50}, //Month
		t7: []byte{49},     //Week day
	}
	expected := SpecScheduleTime{
		Minute:  "31",
		Hour:    "2",
		Day:     "4",
		Month:   "12",
		WeekDay: "1",
	}
	actual := toSpecScheduleTime(args)
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v, got %v\n", expected, actual)
	}
}
