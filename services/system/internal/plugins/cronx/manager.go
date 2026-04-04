package cronx

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"wklive/services/system/models"

	"github.com/robfig/cron/v3"
)

type JobHandler func(ctx context.Context, job *models.SysJob) error

type CronManager struct {
	cron        *cron.Cron
	mu          sync.RWMutex
	entryMap    map[int64]cron.EntryID   // jobID -> cron entry id
	jobMap      map[int64]*models.SysJob // jobID -> job
	handlerMap  map[string]JobHandler    // invokeTarget -> handler
	runningMap  sync.Map                 // jobID -> struct{}
	cancelMap   sync.Map                 // jobID -> context.CancelFunc
	jobLogModel models.JobLogModel       // job log model
}

// NewCronManager
// 支持 5 位和 6 位 cron：
// 5位：*/5 * * * *
// 6位：0 */5 * * * *
func NewCronManager(jobLogModel models.JobLogModel) *CronManager {
	parser := cron.NewParser(
		cron.SecondOptional |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow |
			cron.Descriptor,
	)

	logger := cron.PrintfLogger(log.New(os.Stdout, "[cron] ", log.LstdFlags))

	c := cron.New(
		cron.WithParser(parser),
		cron.WithChain(cron.Recover(logger)),
	)

	return &CronManager{
		cron:        c,
		entryMap:    make(map[int64]cron.EntryID),
		jobMap:      make(map[int64]*models.SysJob),
		handlerMap:  make(map[string]JobHandler),
		jobLogModel: jobLogModel,
	}
}

// StartScheduler 启动整个调度器
func (m *CronManager) StartScheduler() {
	m.cron.Start()
}

// StopScheduler 停止整个调度器（等待当前任务自然结束）
// 注意：这不会强杀正在执行的任务
func (m *CronManager) StopScheduler() context.Context {
	return m.cron.Stop()
}

// RegisterHandler 注册任务执行器
func (m *CronManager) RegisterHandler(invokeTarget string, handler JobHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlerMap[invokeTarget] = handler
}

// StartJob 启动一个任务（加入调度）
// 如果已存在，会先移除旧任务再重新注册
func (m *CronManager) StartJob(job *models.SysJob) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}
	if job.Id <= 0 {
		return fmt.Errorf("job id is invalid")
	}
	if job.CronExpression == "" {
		return fmt.Errorf("job[%d] cron expression is empty", job.Id)
	}
	if job.InvokeTarget == "" {
		return fmt.Errorf("job[%d] invoke target is empty", job.Id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	handler, ok := m.handlerMap[job.InvokeTarget]
	if !ok {
		return fmt.Errorf("job[%d] handler not found: %s", job.Id, job.InvokeTarget)
	}

	// 已存在先移除
	if oldEntryID, exists := m.entryMap[job.Id]; exists {
		m.cron.Remove(oldEntryID)
		delete(m.entryMap, job.Id)
	}

	// 拷贝一份，避免外部修改影响运行时
	jobCopy := *job

	entryID, err := m.cron.AddFunc(job.CronExpression, func() {
		m.execute(&jobCopy, handler)
	})
	if err != nil {
		return fmt.Errorf("job[%d] add cron failed: %w", job.Id, err)
	}

	m.entryMap[job.Id] = entryID
	m.jobMap[job.Id] = &jobCopy
	return nil
}

// PauseJob 暂停一个任务（只停止后续调度，不强杀当前执行）
func (m *CronManager) PauseJob(jobID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entryID, ok := m.entryMap[jobID]; ok {
		m.cron.Remove(entryID)
		delete(m.entryMap, jobID)
	}
}

// RemoveJob 删除运行中的任务注册信息
func (m *CronManager) RemoveJob(jobID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entryID, ok := m.entryMap[jobID]; ok {
		m.cron.Remove(entryID)
		delete(m.entryMap, jobID)
	}
	delete(m.jobMap, jobID)
}

// ReloadJob 重载任务
// 常用于修改 cron_expression / invoke_target 后重新生效
func (m *CronManager) ReloadJob(job *models.SysJob) error {
	m.PauseJob(job.Id)
	if job.Status != 1 {
		return nil
	}
	return m.StartJob(job)
}

// RunOnce 立即执行一次，不依赖 cron 调度
func (m *CronManager) RunOnce(job *models.SysJob) error {
	if job == nil {
		return fmt.Errorf("job is nil")
	}

	m.mu.RLock()
	handler, ok := m.handlerMap[job.InvokeTarget]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("job[%d] handler not found: %s", job.Id, job.InvokeTarget)
	}

	return m.execute(job, handler)
}

// StopRunningJob 停掉当前正在执行的任务
// 前提：你的 handler 里要监听 ctx.Done()
func (m *CronManager) StopRunningJob(jobID int64) bool {
	v, ok := m.cancelMap.Load(jobID)
	if !ok {
		return false
	}
	cancel, ok := v.(context.CancelFunc)
	if !ok {
		return false
	}
	cancel()
	return true
}

// IsStarted 判断任务是否已经加入调度
func (m *CronManager) IsStarted(jobID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.entryMap[jobID]
	return ok
}

// IsRunning 判断任务当前是否在执行中
func (m *CronManager) IsRunning(jobID int64) bool {
	_, ok := m.runningMap.Load(jobID)
	return ok
}

// LoadJobs 批量加载任务（一般服务启动时调用）
func (m *CronManager) LoadJobs(jobs []*models.SysJob) error {
	for _, job := range jobs {
		if job == nil {
			continue
		}
		if job.Status != 1 {
			continue
		}
		if err := m.StartJob(job); err != nil {
			return err
		}
	}
	return nil
}

// ListStartedJobIDs 查看当前已启动的任务ID
func (m *CronManager) ListStartedJobIDs() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]int64, 0, len(m.entryMap))
	for jobID := range m.entryMap {
		ids = append(ids, jobID)
	}
	return ids
}

// execute 实际执行任务
// 默认同一个任务不允许并发执行
func (m *CronManager) execute(job *models.SysJob, handler JobHandler) error {
	startTime := time.Now().UnixMilli()
	if _, loaded := m.runningMap.LoadOrStore(job.Id, struct{}{}); loaded {
		log.Printf("job[%d-%s] is already running, skip this time", job.Id, job.JobName)
		return nil
	}
	defer m.runningMap.Delete(job.Id)

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelMap.Store(job.Id, cancel)
	defer func() {
		cancel()
		m.cancelMap.Delete(job.Id)
	}()

	log.Printf("job start: id=%d, name=%s, target=%s", job.Id, job.JobName, job.InvokeTarget)
	err := handler(ctx, job)
	endTime := time.Now().UnixMilli()
	status := int64(1)
	message := "success"
	exceptionInfo := ""
	if err != nil {
		log.Printf("job failed: id=%d, name=%s, err=%v", job.Id, job.JobName, err)
		status = 0
		message = "failed"
		exceptionInfo = err.Error()
	} else {
		log.Printf("job success: id=%d, name=%s", job.Id, job.JobName)
	}
	_, err = m.jobLogModel.Insert(context.Background(), &models.SysJobLog{
		JobId:          job.Id,
		JobName:        job.JobName,
		InvokeTarget:   job.InvokeTarget,
		CronExpression: sql.NullString{String: job.CronExpression},
		Status:         status,
		Message:        sql.NullString{String: message, Valid: true},
		ExceptionInfo:  sql.NullString{String: exceptionInfo},
		StartTime:      startTime,
		EndTime:        endTime,
		CreateTimes:    time.Now().UnixMilli(),
	})
	return err
}

func (m *CronManager) LoadRegisteredHandlers() {
	handlers := GetRegisteredHandlers()
	for name, handler := range handlers {
		m.RegisterHandler(name, handler)
	}
}
