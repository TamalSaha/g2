package client

import (
	rt "github.com/appscode/g2/pkg/runtime"
)

type SchedTimeWithData struct {
	rt.SpecScheduleTime
	data       string
}

func (this SchedTimeWithData) Bytes() []byte {
	sts := this.SpecScheduleTime.Bytes()
	a := len(sts)
	b := len(this.data)
	l := a+b
	data := rt.NewBuffer(l)
	copy(data[0:a], sts)
	copy(data[a:], this.data)
	return data
}
