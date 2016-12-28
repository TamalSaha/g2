package client

import (
	"testing"

	rt "github.com/appscode/g2/pkg/runtime"
	"fmt"
	"sync"
	"os"
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

func TestTest(t *testing.T) {
	c, _ := New("tcp4", "127.0.0.1:4730")
	// ... error handling
	defer c.Close()
	c.ErrorHandler = func(e error) {
		fmt.Println(e)
	}

	//echo := []byte("Hello world")
	//echomsg, _ := c.Echo(echo)
	//fmt.Println(string(echomsg))

	//echo = []byte("Hello\x00 world")
	//echomsg, _ = c.Echo(echo)
	//fmt.Println(string(echomsg))

	//time.Sleep(30*time.Second)

	//handeler := func(resp * Response) {
	//	switch resp.DataType {
	//	case rt.PT_WorkException:
	//		fallthrough
	//	case rt.PT_WorkFail:
	//		fallthrough
	//	case rt.PT_WorkComplete:
	//		if data, err := resp.Result(); err == nil {
	//			fmt.Printf("RESULT: %v\n", data)
	//		} else {
	//			fmt.Printf("RESULT: %s\n", err)
	//		}
	//	case rt.PT_WorkWarning:
	//		fallthrough
	//	case rt.PT_WorkData:
	//		if data, err := resp.Update(); err == nil {
	//			fmt.Printf("UPDATE: %v\n", data)
	//		} else {
	//			fmt.Printf("UPDATE: %v, %s\n", data, err)
	//		}
	//	case rt.PT_WorkStatus:
	//		if data, err := resp.Status(); err == nil {
	//			fmt.Printf("STATUS: %v\n", data)
	//		} else {
	//			fmt.Printf("STATUS: %s\n", err)
	//		}
	//	default:
	//		fmt.Printf("UNKNOWN: %v", resp.Data)
	//	}
	//}
	//_, err := c.Do("Test", echo, rt.JobLow, handeler)
	//handle, err := c.DoBg("Test", echo, rt.JobLow)


	handle, err := c.DoSched("Test", rt.SpecScheduleTime{
		Minute: "50,48,49",
		Hour: "*",
		Dom: "*",
		Month: "*",
		Dow: "*",
	})

	fmt.Print("Handle: ", handle)
	//handle, err := c.DoSched("Test", []byte("1234"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//status, err := c.Status(handle)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
	//fmt.Printf("%v", *status)
	//
	//_, err = c.DoSched("Foobar", echo, jobHandler)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(0)
	//}

	fmt.Println("Press Ctrl-C to exit ...")
	var mutex sync.Mutex
	mutex.Lock()
	mutex.Lock()
}
func TestClientEcho(t *testing.T) {
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
	handle, err := client.DoBg("ToUpper", []byte("abcdef"), rt.JobLow)
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
