package runtime

import (
	"fmt"
	"github.com/appscode/errors"
	"strconv"
	"strings"
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

type specScheduleTime struct {
	Minute   string `json:"minute"`
	Hour     string `json:"hour"`
	Day      string `json:"day"`
	Month    string `json:"month"`
	WeekDay  string `json:"week_day"`
	cronExpr string
}

func NewSchedule(minute, hour, dayOfMonth, month, dayOfWeek string) (specScheduleTime, error) {
	replaceSpaceWithStar := func(t string) string {
		if strings.TrimSpace(t) == "" {
			return "*"
		}
		return t
	}
	minute = replaceSpaceWithStar(minute)
	hour = replaceSpaceWithStar(hour)
	dayOfMonth = replaceSpaceWithStar(dayOfMonth)
	month = replaceSpaceWithStar(month)
	dayOfWeek = replaceSpaceWithStar(dayOfWeek)

	if !(validateTimeUnit(minute, 0, 59) &&
		validateTimeUnit(hour, 0, 23) &&
		validateTimeUnit(dayOfMonth, 1, 31) &&
		validateTimeUnit(month, 1, 12) &&
		validateTimeUnit(dayOfWeek, 0, 6)) {

		return specScheduleTime{}, errors.NewGoError("invalid cron expression")
	}

	scdt := specScheduleTime{
		Minute:  minute,
		Hour:    hour,
		Day:     dayOfMonth,
		Month:   month,
		WeekDay: dayOfWeek,
	}
	scdt.cronExpr = fmt.Sprintf("%v %v %v %v %v", scdt.Minute, scdt.Hour, scdt.Day, scdt.Month, scdt.WeekDay)
	return scdt, nil
}

func NewScheduleFromExpression(exp string) (specScheduleTime, error) {
	toks := strings.Split(exp, " ")
	if len(toks) != 5 {
		return specScheduleTime{}, errors.NewGoError("invalid cron expression")
	}
	return NewSchedule(toks[0], toks[1], toks[2], toks[3], toks[4])
}

func (self specScheduleTime) Bytes() []byte {
	fix := func(v string) string {
		if v == "*" {
			return ""
		}
		return v
	}
	return []byte(fmt.Sprintf("%v\x00%v\x00%v\x00%v\x00%v\x00", fix(self.Minute), fix(self.Hour), fix(self.Day), fix(self.Month), fix(self.WeekDay)))
}

func (self specScheduleTime) CronExpr() string {
	return self.cronExpr
}

func validateTimeUnit(t string, st, end int) bool {
	if t == "*" {
		return true
	}
	v, err := strconv.Atoi(t)
	if err != nil {
		return false
	}
	if v >= st && v <= end {
		return true
	}
	return false
}
