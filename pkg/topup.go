package pkg

import (
	"context"
	"time"
)

type Topup struct {
	ctx      context.Context
	cancel   context.CancelFunc
	batchId  string
	interval time.Duration
}

func NewTopupTask(ctx context.Context, batchId string, interval time.Duration) *Topup {
	ctx2, cancel := context.WithCancel(ctx)
	return &Topup{
		batchId:  batchId,
		interval: interval,
		ctx:      ctx2,
		cancel:   cancel,
	}
}

func (t *Topup) Execute(context.Context) error {
	for {
		select {
		case <-time.After(t.interval):
			// TODO
		case <-t.ctx.Done():
			return nil
		}
	}
}

func (t *Topup) Name() string {
	return t.batchId
}

func (t *Topup) Stop() {
	t.cancel()
}
