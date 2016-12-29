package storage

import (
	. "github.com/appscode/g2/pkg/runtime"
)

type Db interface {
	JobQueue
	SchedJobQueue
}

type JobQueue interface {
	AddJob(j *Job) error
	DoneJob(j *Job) error
	GetJobs() ([]*Job, error)
}

type SchedJobQueue interface {
	AddShedJob(sj *ScheduledJob) error
	DeleteSchedJob(sj *ScheduledJob) error
	GetShedJobs() ([]*ScheduledJob, error)
}
