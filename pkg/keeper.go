package pkg

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/plexsysio/taskmanager"
)

type Keeper struct {
	ctx         context.Context
	cancel      context.CancelFunc
	taskManager *taskmanager.TaskManager
	url         string

	tasks map[string]*Topup
	mtx   sync.Mutex
}

func New(ctx context.Context, url string) *Keeper {
	ctx2, cancel := context.WithCancel(ctx)
	return &Keeper{
		ctx:         ctx2,
		cancel:      cancel,
		taskManager: taskmanager.New(1, 100, time.Second*15),
		url:         url,
		tasks:       map[string]*Topup{},
	}
}

func (k *Keeper) Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string) error {

	if _, err := strconv.ParseInt(minBalance, 10, 64); err != nil {
		return err
	}
	if _, err := strconv.ParseInt(topupBalance, 10, 64); err != nil {
		return err
	}

	minAmount := &big.Int{}
	minAmount.SetString(minBalance, 10)

	topAmount := &big.Int{}
	topAmount.SetString(topupBalance, 10)

	intervalDuration, err := time.ParseDuration(interval)
	if err != nil {
		return err
	}

	task, err := newTopupTask(k.ctx, name, batchId, k.url, balanceEndpoint, minAmount, topAmount, intervalDuration)
	if err != nil {
		return err
	}
	started, err := k.taskManager.Go(task)
	if err != nil {
		return err
	}
	<-started
	k.mtx.Lock()
	defer k.mtx.Unlock()
	k.tasks[batchId] = task
	return nil
}

func (k *Keeper) Unwatch(batchId string) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	task := k.tasks[batchId]
	task.Stop()
	delete(k.tasks, batchId)
	return nil
}

func (k *Keeper) List() []string {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	tasks := []string{}
	for i, v := range k.tasks {
		fmt.Println(v.Name())
		tasks = append(tasks, i)
	}
	return tasks
}

func (k *Keeper) Stop() {
	k.cancel()
}
