package klinewriter

import (
	"context"
	"fmt"
	"sync"
	"time"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type batchKey struct {
	CategoryCode string
	Market       string
	Interval     string
}

type BatchWriter struct {
	factory *models.CoinKlineModelFactory

	batchSize     int
	flushInterval time.Duration
	writeTimeout  time.Duration

	ch      chan *models.CoinKline
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	buffers map[batchKey][]*models.CoinKline
}

func NewBatchWriter(
	factory *models.CoinKlineModelFactory,
	queueSize int,
	batchSize int,
	flushInterval time.Duration,
	writeTimeout time.Duration,
) *BatchWriter {
	if queueSize <= 0 {
		queueSize = 10000
	}
	if batchSize <= 0 {
		batchSize = 200
	}
	if flushInterval <= 0 {
		flushInterval = 1 * time.Second
	}
	if writeTimeout <= 0 {
		writeTimeout = 5 * time.Second
	}

	return &BatchWriter{
		factory:       factory,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		writeTimeout:  writeTimeout,
		ch:            make(chan *models.CoinKline, queueSize),
		stopCh:        make(chan struct{}),
		buffers:       make(map[batchKey][]*models.CoinKline),
	}
}

func (w *BatchWriter) Start() {
	w.wg.Add(1)
	go w.run()
}

func (w *BatchWriter) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

func (w *BatchWriter) Enqueue(data *models.CoinKline) error {
	if data == nil {
		return fmt.Errorf("coin kline is nil")
	}

	select {
	case w.ch <- data:
		return nil
	default:
		return fmt.Errorf("batch writer queue full")
	}
}

func (w *BatchWriter) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data := <-w.ch:
			if data == nil {
				continue
			}
			w.add(data)

		case <-ticker.C:
			w.flushAll()

		case <-w.stopCh:
			w.drain()
			w.flushAll()
			return
		}
	}
}

func (w *BatchWriter) add(data *models.CoinKline) {
	key := batchKey{
		CategoryCode: data.CategoryCode,
		Market:       data.Market,
		Interval:     data.Interval,
	}

	w.mu.Lock()
	w.buffers[key] = append(w.buffers[key], data)
	size := len(w.buffers[key])

	var toFlush []*models.CoinKline
	if size >= w.batchSize {
		toFlush = w.buffers[key]
		w.buffers[key] = nil
	}
	w.mu.Unlock()

	if len(toFlush) > 0 {
		w.flush(key, toFlush)
	}
}

func (w *BatchWriter) flushAll() {
	w.mu.Lock()
	snapshot := make(map[batchKey][]*models.CoinKline, len(w.buffers))
	for k, v := range w.buffers {
		if len(v) > 0 {
			snapshot[k] = v
		}
	}
	w.buffers = make(map[batchKey][]*models.CoinKline)
	w.mu.Unlock()

	for k, list := range snapshot {
		w.flush(k, list)
	}
}

func (w *BatchWriter) drain() {
	for {
		select {
		case data := <-w.ch:
			if data != nil {
				w.add(data)
			}
		default:
			return
		}
	}
}

func (w *BatchWriter) flush(key batchKey, list []*models.CoinKline) {
	if len(list) == 0 {
		return
	}

	model := w.factory.New(key.CategoryCode, key.Interval)
	if model == nil {
		logx.Errorf("batch flush skipped, invalid model, categoryCode=%s interval=%s size=%d",
			key.CategoryCode, key.Interval, len(list))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), w.writeTimeout)
	defer cancel()

	err := model.BulkUpsertBySymbolTs(ctx, list)
	if err != nil {
		logx.Errorf("batch bulk upsert error, categoryCode=%s interval=%s size=%d err=%v",
			key.CategoryCode, key.Interval, len(list), err)
		return
	}

	logx.Infof("batch bulk upsert success, categoryCode=%s interval=%s size=%d",
		key.CategoryCode, key.Interval, len(list))
}
