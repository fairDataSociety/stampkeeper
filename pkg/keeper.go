/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
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
	logger      Logger
	tasks       map[string]*Topup
	mtx         sync.Mutex
}

func New(ctx context.Context, url string, logger Logger) *Keeper {
	ctx2, cancel := context.WithCancel(ctx)
	return &Keeper{
		ctx:         ctx2,
		cancel:      cancel,
		taskManager: taskmanager.New(1, 100, time.Second*15, logger),
		url:         url,
		logger:      logger,
		tasks:       map[string]*Topup{},
	}
}

func (k *Keeper) Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string, cb func(action *TopupAction) error) error {

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

	task, err := newTopupTask(k.ctx, name, batchId, k.url, balanceEndpoint, minAmount, topAmount, intervalDuration, cb, k.logger)
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
	if k.tasks[batchId] != nil {
		task := k.tasks[batchId]
		task.active = false
		task.Stop()
		return nil
	}
	return fmt.Errorf("stampkeeper not running for this batch id")
}

func (k *Keeper) List() []interface{} {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	tasks := []interface{}{}
	for i, v := range k.tasks {
		info := map[string]interface{}{}
		info["batch"] = i
		info["active"] = v.active
		tasks = append(tasks, info)
	}
	return tasks
}

func (k *Keeper) GetTaskInfo(batchId string) (map[string]interface{}, error) {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	info := map[string]interface{}{}
	if k.tasks[batchId] != nil {
		info["batch"] = batchId
		return info, nil
	}
	return nil, fmt.Errorf("stampkeeper not running for this batch id")
}

func (k *Keeper) Stop() {
	k.cancel()
}
