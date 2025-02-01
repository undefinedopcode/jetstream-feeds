package main

import (
	"context"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

const maxQueueSize = 256

type WorkItem struct {
	name             string
	seq              int
	payload          any
	retriesAvailable int
	retryAfter       time.Time
	attempts         int
}

type WorkItemResult struct {
	item   *WorkItem
	reason error
}

type WorkHandler func(job *WorkItem) (error, bool)
type RetryBackoffFunc func(attempts int) time.Duration

var dummyBackoffFunc = func(attempt int) time.Duration {
	const baseDelay = 5 * time.Second
	var multiple = 2 << attempt
	return time.Duration(multiple) * baseDelay
}

type Worker struct {
	name           string
	work           chan *WorkItem
	retriable      chan *WorkItem
	maxRetries     int
	handler        WorkHandler
	dlq            chan *WorkItemResult
	useDlq         bool
	maxConcurrency int
	seq            int
	ctx            context.Context
	ctxCancel      context.CancelFunc
	backoff        RetryBackoffFunc
	logger         *log.Logger
	sync.Mutex
}

func NewWorker(name string, handler WorkHandler, maxRetries int, maxConcurrency int, useDLQ bool, backoff RetryBackoffFunc, logger *log.Logger) *Worker {
	ctx, ctxCancel := context.WithCancel(context.Background())
	if logger == nil {
		logger = log.Default()
	}
	w := &Worker{
		name:           name,
		handler:        handler,
		work:           make(chan *WorkItem, maxQueueSize),
		retriable:      make(chan *WorkItem, maxQueueSize),
		maxRetries:     maxRetries,
		maxConcurrency: maxConcurrency,
		dlq:            make(chan *WorkItemResult, maxQueueSize),
		useDlq:         useDLQ,
		ctx:            ctx,
		ctxCancel:      ctxCancel,
		backoff:        backoff,
		logger:         logger,
	}
	return w
}

func (w *Worker) getSeq() int {
	w.Lock()
	defer w.Unlock()
	w.seq++
	return w.seq
}

func (w *Worker) runner(ctx context.Context, idNum int) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker stopping", "worker", w.name, "idnum", idNum, "reason", ctx.Err())
			return
		case wi := <-w.work:
			if wi != nil {
				if w.handler != nil {
					err, isFatal := w.handler(wi)
					if err != nil {
						w.logger.Error("Worker failed processing job", "worker", w.name, "job_id", wi.seq, "error", err)
						if isFatal || wi.retriesAvailable < 1 {
							if w.useDlq {
								w.dlq <- &WorkItemResult{
									item:   wi,
									reason: err,
								}
							}
						} else {
							f := w.backoff
							if f == nil {
								f = dummyBackoffFunc
							}
							delay := f(wi.attempts)
							wi.attempts++
							wi.retryAfter = time.Now().Add(delay)
							wi.retriesAvailable--
							w.retriable <- wi
							w.logger.Info("Worker scheduled retry for job", "worker", w.name, "job_id", wi.seq, "retry_in", delay)
						}
					}
				}
			}
		case wi := <-w.retriable:
			if wi != nil {
				if !time.Now().Before(wi.retryAfter) {
					if w.handler != nil {
						err, isFatal := w.handler(wi)
						if err != nil {
							w.logger.Error("Worker failed processing job on retry", "worker", w.name, "job_id", wi.seq, "error", err)
							if isFatal || wi.retriesAvailable < 1 {
								if w.useDlq {
									w.dlq <- &WorkItemResult{
										item:   wi,
										reason: err,
									}
								}
							} else {
								f := w.backoff
								if f == nil {
									f = dummyBackoffFunc
								}
								delay := f(wi.attempts)
								wi.attempts++
								wi.retryAfter = time.Now().Add(delay)
								wi.retriesAvailable--
								w.retriable <- wi
								w.logger.Info("Worker scheduled retry for job", "worker", w.name, "job_ud", wi.seq, "retry_in", delay)
							}
						}
					}
				} else {
					if w.retriable != nil {
						w.retriable <- wi
						time.Sleep(10 * time.Millisecond)
					}
				}
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (w *Worker) Start() {
	for i := 0; i < w.maxConcurrency; i++ {
		ctx, _ := context.WithCancel(w.ctx)
		ii := i
		go w.runner(ctx, ii)
		w.logger.Info("Worker starting", "worker", w.name, "id", i)
	}
}

func (w *Worker) Stop() {
	w.ctxCancel()
	workRemaining := len(w.work)
	retriesRemaining := len(w.retriable)
	dlqRemaining := len(w.dlq)
	if workRemaining+retriesRemaining+dlqRemaining > 0 {
		log.Warn("Worker has remaining in flight work at shutdown", "worker", w.name, "new_items", workRemaining, "retry_items", retriesRemaining, "dlq_items", dlqRemaining)
	}
	close(w.work)
	close(w.retriable)
	close(w.dlq)
}

func (w *Worker) AddWork(payload any) {
	w.work <- &WorkItem{
		name:             w.name,
		seq:              w.getSeq(),
		payload:          payload,
		retriesAvailable: w.maxRetries,
		retryAfter:       time.Now(),
	}
}
