package server

import "gopkg.in/robfig/cron.v2"

type CronService struct {
	*cron.Cron
}

func NewCronService() *CronService {
	return &CronService{
		Cron: cron.New(),
	}
}

func (self *CronService) Start(){
	self.Cron.Start()
}

func (self *CronService) Stop() {
	self.Cron.Stop()
}

func (self *CronService) AddJob() {
}

func (self *CronService) RemoveJob() {
}