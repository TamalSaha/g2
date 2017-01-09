package client

import (
	"fmt"
	rt "github.com/appscode/g2/pkg/runtime"

)

type ScheduleTimedData struct {
	cronSchedule rt.CronSpecInterface
	data         []byte
}

func (this ScheduleTimedData) Bytes() []byte {
	return []byte(fmt.Sprintf("%v%v", string(this.cronSchedule.Bytes()), string(this.data)))
}