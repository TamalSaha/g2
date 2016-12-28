package runtime

import (
	"encoding/json"
	"bytes"
	"fmt"
)
/*
SUBMIT_JOB_SCHED

    Just like SUBMIT_JOB_BG, but run job at given time instead of
    immediately.

    Arguments:
    - NULL byte terminated function name.
    - NULL byte terminated unique ID.
    - NULL byte terminated minute (0-59).
    - NULL byte terminated hour (0-23).
    - NULL byte terminated day of month (1-31).
    - NULL byte terminated month (1-12).
    - NULL byte terminated day of week (0-6, 0 = Monday).
    - Opaque data that is given to the function as an argument.
 */

type SpecScheduleTime struct {
	Minute string
	Hour   string
	Dom    string
	Month  string
	Dow    string
}

func NewSchedTime(minute, hour, dayOfMonth, month, dayOfWeek string) SpecScheduleTime {
	return SpecScheduleTime{
		Minute: minute,
		Hour: hour,
		Dom: dayOfMonth,
		Month: month,
		Dow: dayOfWeek,
	}
}

func (self SpecScheduleTime) Bytes() []byte {
	a := len(self.Minute)
	b := len(self.Hour)
	c := len(self.Dom)
	d := len(self.Month)
	e := len(self.Dow)
	l := a+b+c+d+e+5
	data := NewBuffer(l)
	copy(data[0:a], self.Minute)
	copy(data[a+1:a+b+1], self.Hour)
	copy(data[a+b+2: a+b+c+2], self.Dom)
	copy(data[a+b+c+3: a+b+c+d+3], self.Month)
	copy(data[a+b+c+d+4: a+b+c+d+e+4], self.Dow)
	return data
}

func (self SpecScheduleTime) ToCronExpr() string  {
	return fmt.Sprintf("%v %v %v %v %v", self.Minute, self.Hour, self.Dom, self.Month, self.Dow)
}

func (self *SpecScheduleTime) String() string {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	m := make(map[string]interface{})
	m["Minute"] = self.Minute
	m["Hour"] = self.Hour
	m["Dom"] = self.Dom
	m["Month"] = self.Month
	m["Dow"] = self.Dow
	if err := enc.Encode(m); err != nil {
		return ""
	}
	return string(b.Bytes())
}