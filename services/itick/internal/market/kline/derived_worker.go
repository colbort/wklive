package kline

import (
	"fmt"
	"sync"

	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type derivedJob struct {
	minutes []*models.CoinKline
	done    chan error
}

// DerivedWorker serializes higher-interval rebuilds so Mongo queries and Redis
// writes never block the 1m BatchWriter consumer.
type DerivedWorker struct {
	aggregator *DerivedAggregator
	jobs       chan derivedJob
	stopOnce   sync.Once
	stopCh     chan struct{}
	doneCh     chan struct{}
}

func NewDerivedWorker(aggregator *DerivedAggregator, queueSize int) *DerivedWorker {
	if queueSize <= 0 {
		queueSize = 1024
	}
	return &DerivedWorker{aggregator: aggregator, jobs: make(chan derivedJob, queueSize),
		stopCh: make(chan struct{}), doneCh: make(chan struct{})}
}

func (w *DerivedWorker) Start() { go w.run() }

func (w *DerivedWorker) Stop() {
	w.stopOnce.Do(func() { close(w.stopCh) })
	<-w.doneCh
}

func (w *DerivedWorker) Enqueue(minutes []*models.CoinKline) error {
	job := derivedJob{minutes: cloneKlines(minutes)}
	select {
	case w.jobs <- job:
		return nil
	default:
		return fmt.Errorf("derived kline queue full")
	}
}

// Rebuild waits for its job and is used by REST flows that must not report
// success before all affected higher intervals are persisted.
func (w *DerivedWorker) Rebuild(minutes []*models.CoinKline) error {
	done := make(chan error, 1)
	job := derivedJob{minutes: cloneKlines(minutes), done: done}
	select {
	case w.jobs <- job:
	case <-w.stopCh:
		return fmt.Errorf("derived kline worker stopped")
	}
	select {
	case err := <-done:
		return err
	case <-w.stopCh:
		return fmt.Errorf("derived kline worker stopped")
	}
}

func (w *DerivedWorker) run() {
	defer close(w.doneCh)
	for {
		select {
		case job := <-w.jobs:
			w.execute(job)
		case <-w.stopCh:
			for {
				select {
				case job := <-w.jobs:
					w.execute(job)
				default:
					return
				}
			}
		}
	}
}

func (w *DerivedWorker) execute(job derivedJob) {
	err := w.aggregator.Rebuild(job.minutes)
	if job.done != nil {
		job.done <- err
		return
	}
	if err != nil {
		logx.Errorf("async derived kline rebuild failed: %v", err)
	}
}

func cloneKlines(list []*models.CoinKline) []*models.CoinKline {
	out := make([]*models.CoinKline, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		copyItem := *item
		out = append(out, &copyItem)
	}
	return out
}
