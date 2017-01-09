package runtime

import (
	"gopkg.in/robfig/cron.v2"
	"github.com/appscode/errors"
	"fmt"
)

const starBit  = 1<<63
type CronSpecInterface interface {
	Bytes() []byte
	Schedule() cron.Schedule
}

type cronSpec struct {
	specByte []byte
	schedule *cron.SpecSchedule
}

func NewCronSchedule(expr string) (CronSpecInterface, error)  {
	scd, err := cron.Parse(expr)
	if err != nil {
		return nil, err
	}
	specScd, ok := scd.(*cron.SpecSchedule)
	if !ok {
		return nil, errors.New("invalid cron expression")
	}
	cronByte := []byte(fmt.Sprintf("%v\x00%v\x00%v\x00%v\x00%v\x00", fix(specScd.Minute), fix(specScd.Hour), fix(specScd.Dom), fix(specScd.Month), fix(specScd.Dow)))
	return cronSpec{
		specByte: cronByte,
		schedule: specScd,
	}, nil

}

func (c cronSpec) Schedule() cron.Schedule  {
	return c.schedule
}

func (c cronSpec) Bytes() []byte  {
	return c.specByte
}

func fix(n uint64) string{
	if hasStar(n) {
		return ""
	}
	var i uint64
	for i = 0; i<64; i++ {
		if((1<<i)&n) != 0 {
			return fmt.Sprintf("%v", i)
		}
	}
	return fmt.Sprintf("%v", i)
}

func hasStar(n uint64) bool {
	return (n&starBit)!=0
}
