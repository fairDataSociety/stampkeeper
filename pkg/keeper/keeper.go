/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package keeper

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/fairDataSociety/stampkeeper/pkg/logging"
	"github.com/fairDataSociety/stampkeeper/pkg/topup"
	"github.com/plexsysio/taskmanager"
)

type Keeper struct {
	ctx         context.Context
	cancel      context.CancelFunc
	taskManager *taskmanager.TaskManager
	url         string
	logger      logging.Logger
	tasks       map[string]*topup.Topup
	mtx         sync.Mutex
}

func New(ctx context.Context, url string, logger logging.Logger) *Keeper {
	ctx2, cancel := context.WithCancel(ctx)
	return &Keeper{
		ctx:         ctx2,
		cancel:      cancel,
		taskManager: taskmanager.New(1, 100, time.Second*15, logger),
		url:         url,
		logger:      logger,
		tasks:       map[string]*topup.Topup{},
	}
}

func (k *Keeper) Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string, cb func(action *topup.TopupAction) error) error {

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

	task, err := topup.NewTopupTask(k.ctx, name, batchId, k.url, balanceEndpoint, minAmount, topAmount, intervalDuration, cb, k.logger)
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
		task.Deactivate()
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
		info["active"] = v.State()
		info["name"] = v.Name()
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
