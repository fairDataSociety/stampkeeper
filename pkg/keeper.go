package pkg

import (
	"context"
	"time"

	"github.com/plexsysio/taskmanager"
)

type Keeper struct {
	ctx         context.Context
	cancel      context.CancelFunc
	taskManager *taskmanager.TaskManager
}

func New(ctx context.Context) *Keeper {
	ctx2, cancel := context.WithCancel(ctx)
	return &Keeper{
		ctx:         ctx2,
		cancel:      cancel,
		taskManager: taskmanager.New(1, 100, time.Second*15),
	}
}

func (k *Keeper) Watch(batchId, interval string) error {
	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}
	task := NewTopupTask(k.ctx, batchId, intervalDuration)
	started, err := k.taskManager.Go(task)
	if err != nil {
		return err
	}
	<-started
	return nil
}

func (k *Keeper) Unwatch(batchId string) error {
	// TODO
	return nil
}

func (k *Keeper) List() []string {
	tasks := []string{}
	for _, v := range k.taskManager.Status() {
		tasks = append(tasks, v.TaskName)
	}
	return tasks
}

func (k *Keeper) Stop() {
	k.cancel()
}
