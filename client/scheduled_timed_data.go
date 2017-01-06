package client

import (
	"fmt"
)

type ScheduleTimedData struct {
	scheduledTimeData []byte
	data              []byte
}

func (this ScheduleTimedData) Bytes() []byte {
	return []byte(fmt.Sprintf("%v%v", string(this.scheduledTimeData), string(this.data)))
}
