package client

import (
	"testing"

	rt "github.com/appscode/g2/pkg/runtime"
)

const (
	TestStr = "Hello world"
)

var client *Client

func TestClientAddServer(t *testing.T) {
	t.Log("Add local server 127.0.0.1:4730")
	var err error
	if client, err = New(rt.Network, "127.0.0.1:4730"); err != nil {
		t.Fatal(err)
	}
	client.ErrorHandler = func(e error) {
		t.Log(e)
	}
}

func TestClientEcho(t *testing.T) {
	initClient(t)
	echo, err := client.Echo([]byte(TestStr))
	if err != nil {
		t.Error(err)
		return
	}
	if string(echo) != TestStr {
		t.Errorf("Echo error, %s expected, %s got", TestStr, echo)
		return
	}
}

func TestClientDoBg(t *testing.T) {
	initClient(t)
	handle, err := client.DoBg("ToUpper", []byte("abcdef"), rt.JobNormal)
	if err != nil {
		t.Error(err)
		return
	}
	if handle == "" {
		t.Error("Handle is empty.")
	} else {
		t.Log(handle)
	}
}

func TestClientDoSched(t *testing.T) {
	initClient(t)
	handle, err := client.DoSched("scheduledJobTest", SchedTimeWithData{
		SpecScheduleTime: rt.SpecScheduleTime{
			Minute:  "",
			Hour:    "",
			Day:     "",
			Month:   "",
			WeekDay: "",
		},
		data: "Test data",
	})
	if err != nil {
		t.Error(err)
		return
	}
	if handle == "" {
		t.Error("Handle is empty.")
	} else {
		t.Log(handle)
	}
}

func TestClientDo(t *testing.T) {
	initClient(t)
	jobHandler := func(job *Response) {
		str := string(job.Data)
		if str == "ABCDEF" {
			t.Log(str)
		} else {
			t.Errorf("Invalid data: %s", job.Data)
		}
		return
	}
	handle, err := client.Do("ToUpper", []byte("abcdef"),
		rt.JobLow, jobHandler)
	if err != nil {
		t.Error(err)
		return
	}
	if handle == "" {
		t.Error("Handle is empty.")
	} else {
		t.Log(handle)
	}
}

func TestClientStatus(t *testing.T) {
	status, err := client.Status("handle not exists")
	if err != nil {
		t.Error(err)
		return
	}
	if status.Known {
		t.Errorf("The job (%s) shouldn't be known.", status.Handle)
		return
	}
	if status.Running {
		t.Errorf("The job (%s) shouldn't be running.", status.Handle)
		return
	}

	handle, err := client.Do("Delay5sec", []byte("abcdef"), rt.JobLow, nil)
	if err != nil {
		t.Error(err)
		return
	}
	status, err = client.Status(handle)
	if err != nil {
		t.Error(err)
		return
	}
	if !status.Known {
		t.Errorf("The job (%s) should be known.", status.Handle)
		return
	}
	if status.Running {
		t.Errorf("The job (%s) shouldn't be running.", status.Handle)
		return
	}
}

func TestClientClose(t *testing.T) {
	if err := client.Close(); err != nil {
		t.Error(err)
	}
}

func initClient(t *testing.T) *Client {
	if client == nil {
		var err error
		client, err = New(rt.Network, "127.0.0.1:4730")
		if err != nil {
			t.Fatal(err)
		}
	}
	return client

}
