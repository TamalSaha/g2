package client

import (
	rt "github.com/appscode/g2/pkg/runtime"
)

type SchedTimeWithData struct {
	rt.SpecScheduleTime
	data       string
}

func NewSchedTimeWithData(minute, hour, dom, month, dow, data string) (SchedTimeWithData, error) {
	return SchedTimeWithData{
		SpecScheduleTime: rt.SpecScheduleTime{
			Minute: minute,
			Hour: hour,
			Dom: dom,
			Month: month,
			Dow: dow,
		},
		data: data,
	}, nil
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

//func (self SchedTimeWithData) specScheduleByte() []byte {
//	min := fmt.Sprint("%v", self.Minute)
//	hour := fmt.Sprintf("%v", self.Hour)
//	dom := fmt.Sprintf("%v", self.Dom)
//	month := fmt.Sprintf("%v", self.Month)
//	dow := fmt.Sprintf("%v", self.Dow)
//
//	a := len(min)
//	b := len(hour)
//	c := len(dom)
//	d := len(month)
//	e := len(dow)
//	l := a+b+c+d+e+5
//	data := rt.NewBuffer(l)
//	copy(data[0:a], min)
//	copy(data[a+1:a+b+1], hour)
//	copy(data[a+b+2: a+b+c+2], dom)
//	copy(data[a+b+c+3: a+b+c+d+3], month)
//	copy(data[a+b+c+d+4: a+b+c+d+e+4], dow)
//	return data
//}

