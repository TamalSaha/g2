package server

import (
	"container/list"
	"encoding/json"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"fmt"
	"github.com/appscode/errors"
	. "github.com/appscode/g2/pkg/runtime"
	"github.com/appscode/g2/pkg/storage"
	"github.com/appscode/log"
	"github.com/ngaut/stats"
	lberror "github.com/syndtr/goleveldb/leveldb/errors"
	"gopkg.in/robfig/cron.v2"
	"strings"
)

type Server struct {
	protoEvtCh     chan *event
	ctrlEvtCh      chan *event
	funcWorker     map[string]*jobworkermap //function worker
	worker         map[int64]*Worker
	client         map[int64]*Client
	jobs           map[string]*Job
	startSessionId int64
	opCounter      map[PT]int64
	store          storage.Db
	forwardReport  int64
	cronSvc        *cron.Cron
}

var ( //const replys, to avoid building it every time
	wakeupReply = constructReply(PT_Noop, nil)
	nojobReply  = constructReply(PT_NoJob, nil)
)

func NewServer(store storage.Db) *Server {
	return &Server{
		funcWorker: make(map[string]*jobworkermap),
		protoEvtCh: make(chan *event, 100),
		ctrlEvtCh:  make(chan *event, 100),
		worker:     make(map[int64]*Worker),
		client:     make(map[int64]*Client),
		jobs:       make(map[string]*Job),
		opCounter:  make(map[PT]int64),
		store:      store,
		cronSvc:    cron.New(),
	}
}

func (self *Server) loadAllJobs() {
	jobs, err := self.store.GetJobs()
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("%+v", jobs)
	for _, j := range jobs {
		j.ProcessBy = 0 //no body handle it now
		j.CreateBy = 0  //clear
		self.doAddJob(j)
	}
}

func (self *Server) loadAllCronJobs() {
	schedJobs, err := self.store.GetCronJobs()
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("load scheduled job: %+v", schedJobs)
	for _, sj := range schedJobs {
		self.doAddCronJob(sj)
	}
}

func (self *Server) Start(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	go self.EvtLoop()
	log.Debug("listening on", addr)

	go registerWebHandler(self)

	if self.cronSvc != nil {
		self.cronSvc.Start()
	}
	//load background jobs from storage
	if self.store != nil {
		self.loadAllJobs()
		self.loadAllCronJobs()
	}

	for {
		conn, err := ln.Accept()
		if err != nil { // handle error
			continue
		}

		session := &session{}
		go session.handleConnection(self, conn)
	}
}

func (self *Server) addWorker(l *list.List, w *Worker) {
	for it := l.Front(); it != nil; it = it.Next() {
		if it.Value.(*Worker).SessionId == w.SessionId {
			log.Warning("already add")
			return
		}
	}

	l.PushBack(w) //add to worker list
}

func (self *Server) removeWorker(l *list.List, sessionId int64) {
	for it := l.Front(); it != nil; it = it.Next() {
		if it.Value.(*Worker).SessionId == sessionId {
			log.Debugf("removeWorker sessionId %d", sessionId)
			l.Remove(it)
			return
		}
	}
}

func (self *Server) removeWorkerBySessionId(sessionId int64) {
	for _, jw := range self.funcWorker {
		self.removeWorker(jw.workers, sessionId)
	}
	delete(self.worker, sessionId)
}

func (self *Server) handleCanDo(funcName string, w *Worker) {
	w.canDo[funcName] = true
	jw := self.getJobWorkPair(funcName)
	self.addWorker(jw.workers, w)
	self.worker[w.SessionId] = w
}

func (self *Server) getJobWorkPair(funcName string) *jobworkermap {
	jw, ok := self.funcWorker[funcName]
	if !ok { //create list
		jw = &jobworkermap{workers: list.New(), jobs: list.New()}
		self.funcWorker[funcName] = jw
	}

	return jw
}

func (self *Server) add2JobWorkerQueue(j *Job) {
	jw := self.getJobWorkPair(j.FuncName)
	jw.jobs.PushBack(j)
}

func (self *Server) doAddJob(j *Job) {
	j.ProcessBy = 0 //nobody handle it right now
	self.add2JobWorkerQueue(j)
	self.jobs[j.Handle] = j
	self.wakeupWorker(j.FuncName)
}

func (self *Server) doAddAndPersistJob(j *Job) {
	// persistent job
	log.Debugf("add job %+v", j)
	if self.store != nil {
		if err := self.store.AddJob(j); err != nil {
			log.Warning(err)
		}
	}
	self.doAddJob(j)
}

func (self *Server) doAddCronJob(sj *CronJob) cron.EntryID {

	if strings.HasPrefix(sj.ScheduleTime, EpochTimePrefix) {
		value, err := strconv.ParseInt(sj.ScheduleTime[len(EpochTimePrefix):], 10, 64)
		if err != nil {
			log.Errorln(err)
		}
		self.doAddEpochJob(sj, value)
	} else {
		scdT, err := NewCronSchedule(sj.ScheduleTime)
		if err != nil {
			log.Errorln(err)
			return cron.EntryID(0)
		}
		return self.cronSvc.Schedule(
			scdT.Schedule(),
			cron.FuncJob(
				func() {
					jb := &Job{
						Handle:       allocJobId(),
						Id:           sj.JobTemplete.Id,
						Data:         sj.JobTemplete.Data,
						CreateAt:     time.Now(),
						CreateBy:     sj.JobTemplete.CreateBy,
						FuncName:     sj.JobTemplete.FuncName,
						Priority:     sj.JobTemplete.Priority,
						IsBackGround: sj.JobTemplete.IsBackGround,
					}
					self.doAddAndPersistJob(jb)
				}))
	}
	return cron.EntryID(0)
}

func (self *Server) doAddEpochJob(cj *CronJob, epoch int64) {
	j := &Job{
		Handle:       allocJobId(),
		Id:           cj.JobTemplete.Id,
		Data:         cj.JobTemplete.Data,
		CreateAt:     time.Now(),
		CreateBy:     cj.JobTemplete.CreateBy,
		FuncName:     cj.JobTemplete.FuncName,
		Priority:     cj.JobTemplete.Priority,
		IsBackGround: cj.JobTemplete.IsBackGround,
	}
	after := epoch - time.Now().UTC().Unix()
	if after < 0 {
		after = 0
	}
	time.AfterFunc(time.Second*time.Duration(after), func() {
		self.doAddAndPersistJob(j)
		err := self.DeleteCronJob(cj)
		if err != nil {
			log.Errorln(err)
		}
	})

}

func (self *Server) popJob(sessionId int64) (j *Job) {
	for funcName, cando := range self.worker[sessionId].canDo {
		if !cando {
			continue
		}

		if wj, ok := self.funcWorker[funcName]; ok {
			if wj.jobs.Len() == 0 {
				continue
			}

			job := wj.jobs.Front()
			wj.jobs.Remove(job)
			j = job.Value.(*Job)
			return
		}
	}

	return
}

func (self *Server) wakeupWorker(funcName string) bool {
	wj, ok := self.funcWorker[funcName]
	if !ok || wj.jobs.Len() == 0 || wj.workers.Len() == 0 {
		return false
	}

	for it := wj.workers.Front(); it != nil; it = it.Next() {
		w := it.Value.(*Worker)
		if w.status != wsSleep {
			continue
		}

		log.Debug("wakeup sessionId", w.SessionId)

		w.Send(wakeupReply)
		return true
	}

	return false
}

func (self *Server) checkAndRemoveJob(tp PT, j *Job) {
	switch tp {
	case PT_WorkComplete, PT_WorkException, PT_WorkFail:
		self.removeJob(j)
	}
}

func (self *Server) removeJob(j *Job) {
	delete(self.jobs, j.Handle)
	delete(self.worker[j.ProcessBy].runningJobs, j.Handle)
	if j.IsBackGround {
		log.Debugf("done job: %v", j.Handle)
		if self.store != nil {
			if err := self.store.DeleteJob(j); err != nil {
				log.Warning(err)
			}
		}
	}
}

func (self *Server) handleCloseSession(e *event) error {
	sessionId := e.fromSessionId
	if w, ok := self.worker[sessionId]; ok {
		if sessionId != w.SessionId {
			log.Fatalf("sessionId not match %d-%d, bug found", sessionId, w.SessionId)
		}
		self.removeWorkerBySessionId(w.SessionId)

		//reschedule these jobs, so other workers can handle it
		for handle, j := range w.runningJobs {
			if handle != j.Handle {
				log.Fatal("handle not match %d-%d", handle, j.Handle)
			}
			j.Running = false
			self.doAddJob(j)
		}
	}
	if c, ok := self.client[sessionId]; ok {
		log.Debug("removeClient sessionId", sessionId)
		delete(self.client, c.SessionId)
	}
	e.result <- true //notify close finish

	return nil
}

func (self *Server) handleGetWorker(e *event) (err error) {
	var buf []byte
	defer func() {
		e.result <- string(buf)
	}()
	cando := e.args.t0.(string)
	log.Debug("get worker", cando)
	if len(cando) == 0 {
		workers := make([]*Worker, 0, len(self.worker))
		for _, v := range self.worker {
			workers = append(workers, v)
		}
		buf, err = json.Marshal(workers)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}

	log.Debugf("%+v", self.funcWorker)
	if jw, ok := self.funcWorker[cando]; ok {
		log.Debug(cando, jw.workers.Len())
		workers := make([]*Worker, 0, jw.workers.Len())
		for it := jw.workers.Front(); it != nil; it = it.Next() {
			workers = append(workers, it.Value.(*Worker))
		}
		buf, err = json.Marshal(workers)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}

	return
}

func (self *Server) handleGetJob(e *event) (err error) {
	log.Debug("get jobs", e.handle)
	var buf []byte
	defer func() {
		e.result <- string(buf)
	}()

	if len(e.handle) == 0 {
		jobs := []*Job{}
		for _, v := range self.jobs {
			jobs = append(jobs, v)
		}
		buf, err = json.Marshal(jobs)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}

	if job, ok := self.jobs[e.handle]; ok {
		buf, err = json.Marshal(job)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}

	return
}

func (self *Server) handleGetCronJob(e *event) (err error) {
	log.Debug("get cronjobs", e.handle)
	var buf []byte
	defer func() {
		e.result <- string(buf)
	}()

	if len(e.handle) == 0 {
		cjs, err := self.store.GetCronJobs()
		if err != nil {
			log.Error(err)
			return err
		}
		buf, err = json.Marshal(cjs)
		if err != nil {
			log.Error(err)
			return err
		}
		return nil
	}
	cj, err := self.store.GetCronJob(e.handle)
	if err != nil {
		log.Error(err)
		return err
	}
	data, err := json.Marshal(cj)
	if err != nil {
		log.Error(err)
		return err
	}
	buf = []byte(data)
	return
}

func (self *Server) handleCtrlEvt(e *event) (err error) {
	//args := e.args
	switch e.tp {
	case ctrlCloseSession:
		return self.handleCloseSession(e)
	case ctrlGetJob:
		return self.handleGetJob(e)
	case ctrlGetWorker:
		return self.handleGetWorker(e)
	case ctrlGetCronJob:
		return self.handleGetCronJob(e)
	default:
		log.Warningf("%s, %d", e.tp, e.tp)
	}

	return nil
}

func (self *Server) handleSubmitJob(e *event) {
	args := e.args
	c := args.t0.(*Client)
	self.client[c.SessionId] = c
	funcName := bytes2str(args.t1)
	j := &Job{
		Handle:       allocJobId(),
		Id:           bytes2str(args.t2),
		Data:         args.t3.([]byte),
		CreateAt:     time.Now(),
		CreateBy:     c.SessionId,
		FuncName:     funcName,
		Priority:     cmd2Priority(e.tp),
		IsBackGround: isBackGround(e.tp),
	}
	//log.Debugf("%v, job handle %v, %s", CmdDescription(e.tp), j.Handle, string(j.Data))
	e.result <- j.Handle
	self.doAddAndPersistJob(j)
}

func (self *Server) handleCronJob(e *event) {
	args := e.args
	c := args.t0.(*Client)
	self.client[c.SessionId] = c
	funcName := bytes2str(args.t1)
	sst, err := NewCronSchedule(fmt.Sprintf("%v %v %v %v %v",
		byte2strWithFixSpace(args.t3),
		byte2strWithFixSpace(args.t4),
		byte2strWithFixSpace(args.t5),
		byte2strWithFixSpace(args.t6),
		byte2strWithFixSpace(args.t7)),
	)
	if err != nil {
		log.Errorln(err)
		return
	}
	sj := &CronJob{
		JobTemplete: Job{
			Id:           bytes2str(args.t2),
			Data:         args.t8.([]byte),
			CreateAt:     time.Now(),
			CreateBy:     c.SessionId,
			FuncName:     funcName,
			Priority:     cmd2Priority(e.tp),
			IsBackGround: true,
		},
		ScheduleTime: sst.Expr(),
	}
	sj.Handle = allocSchedJobId()
	e.result <- sj.Handle
	// persistent Cron Job
	log.Debugf("add scheduled job %+v", sj)
	id := self.doAddCronJob(sj)
	if err != nil {
		log.Errorln(err)
	}
	sj.CronEntryID = int(id)
	if self.store != nil {
		if err := self.store.AddCronJob(sj); err != nil {
			log.Errorln(err)
		}
	}
	log.Debugf("Scheduled cron job added with function name `%s`, data '%s' and cron SpecScheduleTime - '%+v'\n", string(sj.JobTemplete.FuncName), string(sj.JobTemplete.Data), sj.ScheduleTime)
}

func (self *Server) handleSubmitEpochJob(e *event) {
	args := e.args
	c := args.t0.(*Client)
	self.client[c.SessionId] = c
	funcName := bytes2str(args.t1)
	epochStr := bytes2str(args.t3)
	val, err := strconv.ParseInt(epochStr, 10, 64)
	if err != nil {
		log.Errorln(err)
		return
	}
	sj := &CronJob{
		JobTemplete: Job{
			Id:           bytes2str(args.t2),
			Data:         args.t4.([]byte),
			CreateAt:     time.Now(),
			CreateBy:     c.SessionId,
			FuncName:     funcName,
			Priority:     cmd2Priority(e.tp),
			IsBackGround: true,
		},
		ScheduleTime: EpochTimePrefix + epochStr,
	}
	sj.Handle = allocSchedJobId()
	e.result <- sj.Handle

	// persistent Cron Job
	log.Debugf("add scheduled epoch job %+v", sj)
	self.doAddEpochJob(sj, val)
	if self.store != nil {
		if err := self.store.AddCronJob(sj); err != nil {
			log.Errorln(err)
		}
	}
}

func (self *Server) handleWorkReport(e *event) {
	args := e.args
	slice := args.t0.([][]byte)
	jobhandle := bytes2str(slice[0])
	sessionId := e.fromSessionId
	j, ok := self.worker[sessionId].runningJobs[jobhandle]

	log.Debugf("%v job handle %v", e.tp, jobhandle)
	if !ok {
		log.Warningf("job information lost, %v job handle %v, %+v",
			e.tp, jobhandle, self.jobs)
		return
	}

	if j.Handle != jobhandle {
		log.Fatal("job handle not match")
	}

	if PT_WorkStatus == e.tp {
		j.Percent, _ = strconv.Atoi(string(slice[1]))
		j.Denominator, _ = strconv.Atoi(string(slice[2]))
	}

	self.checkAndRemoveJob(e.tp, j)

	//the client is not updated with status or notified when the job has completed (it is detached)
	if j.IsBackGround {
		return
	}

	//broadcast all clients, which is a really bad idea
	//for _, c := range self.client {
	//	reply := constructReply(e.tp, slice)
	//	c.Send(reply)
	//}

	//just send to original client, which is a bad idea too.
	//if need work status notification, you should create co-worker.
	//let worker send status to this co-worker
	c, ok := self.client[j.CreateBy]
	if !ok {
		log.Debug(j.Handle, "sessionId", j.CreateBy, "missing")
		return
	}

	reply := constructReply(e.tp, slice)
	c.Send(reply)
	self.forwardReport++
}

func (self *Server) handleProtoEvt(e *event) {
	args := e.args
	if e.tp < ctrlCloseSession {
		self.opCounter[e.tp]++
	}

	if e.tp >= ctrlCloseSession {
		self.handleCtrlEvt(e)
		return
	}
	switch e.tp {
	case PT_CanDo:
		w := args.t0.(*Worker)
		funcName := args.t1.(string)
		self.handleCanDo(funcName, w)
	case PT_CantDo:
		sessionId := e.fromSessionId
		funcName := args.t0.(string)
		if jw, ok := self.funcWorker[funcName]; ok {
			self.removeWorker(jw.workers, sessionId)
		}
		delete(self.worker[sessionId].canDo, funcName)
	case PT_SetClientId:
		w := args.t0.(*Worker)
		w.workerId = args.t1.(string)
	case PT_CanDoTimeout: //todo: fix timeout support, now just as CAN_DO
		w := args.t0.(*Worker)
		funcName := args.t1.(string)
		self.handleCanDo(funcName, w)
	case PT_GrabJobUniq:
		sessionId := e.fromSessionId
		w, ok := self.worker[sessionId]
		if !ok {
			log.Fatalf("unregister worker, sessionId %d", sessionId)
			break
		}

		w.status = wsRunning

		j := self.popJob(sessionId)
		if j != nil {
			j.ProcessAt = time.Now()
			j.ProcessBy = sessionId
			//track this job
			j.Running = true
			w.runningJobs[j.Handle] = j
		} else { //no job
			w.status = wsPrepareForSleep
		}
		//send job back
		e.result <- j
	case PT_PreSleep:
		sessionId := e.fromSessionId
		w, ok := self.worker[sessionId]
		if !ok {
			log.Warningf("unregister worker, sessionId %d", sessionId)
			w = args.t0.(*Worker)
			self.worker[w.SessionId] = w
			break
		}

		w.status = wsSleep
		log.Debugf("worker sessionId %d sleep", sessionId)
		//check if there are any jobs for this worker
		for k := range w.canDo {
			if self.wakeupWorker(k) {
				break
			}
		}
	case PT_SubmitJobLow, PT_SubmitJob, PT_SubmitJobHigh, PT_SubmitJobLowBG, PT_SubmitJobBG, PT_SubmitJobHighBG:
		self.handleSubmitJob(e)
	case PT_SubmitJobSched:
		self.handleCronJob(e)
	case PT_SubmitJobEpoch:
		self.handleSubmitEpochJob(e)
	case PT_GetStatus:
		jobhandle := bytes2str(args.t0)
		if job, ok := self.jobs[jobhandle]; ok {
			e.result <- &Tuple{t0: args.t0, t1: true, t2: job.Running,
				t3: job.Percent, t4: job.Denominator}
			break
		}

		e.result <- &Tuple{t0: args.t0, t1: false, t2: false,
			t3: 0, t4: 100} //always set Denominator to 100 if no status update
	case PT_WorkData, PT_WorkWarning, PT_WorkStatus, PT_WorkComplete,
		PT_WorkFail, PT_WorkException:
		self.handleWorkReport(e)
	default:
		log.Warningf("%s, %d", e.tp, e.tp)
	}
}

func (self *Server) wakeupTravel() {
	for k, jw := range self.funcWorker {
		if jw.jobs.Len() > 0 {
			self.wakeupWorker(k)
		}
	}
}

func (self *Server) pubCounter() {
	for k, v := range self.opCounter {
		stats.PubInt64(k.String(), v)
	}
}

func (self *Server) EvtLoop() {
	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case e := <-self.protoEvtCh:
			self.handleProtoEvt(e)
		case e := <-self.ctrlEvtCh:
			self.handleCtrlEvt(e)
		case <-tick.C:
			self.pubCounter()
			stats.PubInt("len(protoEvtCh)", len(self.protoEvtCh))
			stats.PubInt("worker count", len(self.worker))
			stats.PubInt("job queue length", len(self.jobs))
			stats.PubInt("queue count", len(self.funcWorker))
			stats.PubInt("client count", len(self.client))
			stats.PubInt64("forwardReport", self.forwardReport)
		}
	}
}

func (self *Server) allocSessionId() int64 {
	return atomic.AddInt64(&self.startSessionId, 1)
}

func (self *Server) DeleteCronJob(cj *CronJob) error {
	sj, err := self.store.DeleteCronJob(cj)
	if err == lberror.ErrNotFound {
		log.Errorf("handle `%v` not found\n", cj.Handle)
		return errors.NewGoError(fmt.Sprintf("handle `%v` not found", cj.Handle))
	}
	if err != nil {
		log.Errorln(err)
		return err
	}
	self.cronSvc.Remove(cron.EntryID(sj.CronEntryID))
	log.Debugf("job `%v` successfully cancelled.\n", cj.Handle)
	return nil
}
