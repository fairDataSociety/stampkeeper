package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestTaskManager(t *testing.T) {
	var mtx sync.Mutex
	stampInfo := &stamp{
		BatchID:     correctBatchId,
		Amount:      initialAmount,
		Utilization: 16,
		Depth:       20,
		BucketDepth: 16,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.String(), "/stamps/topup/") {
			mtx.Lock()
			amount := &big.Int{}
			amount.SetString(stampInfo.Amount, 10)
			amount = amount.Add(amount, big.NewInt(10000000))
			stampInfo.Amount = amount.String()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			_ = json.NewEncoder(w).Encode(&mockResponse{BatchID: stampInfo.BatchID})
			mtx.Unlock()
		} else if strings.HasPrefix(r.URL.String(), "/stamps/dilute/") {
			mtx.Lock()
			amount := &big.Int{}
			amount.SetString(stampInfo.Amount, 10)
			amount = amount.Sub(amount, big.NewInt(5000000))
			stampInfo.Amount = amount.String()
			stampInfo.Depth = stampInfo.Depth + 2
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			_ = json.NewEncoder(w).Encode(&mockResponse{BatchID: stampInfo.BatchID})
			mtx.Unlock()
		} else if strings.HasPrefix(r.URL.String(), "/stamps/") {
			mtx.Lock()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(stampInfo)
			mtx.Unlock()
		} else {
			fmt.Println()
		}
	}))
	defer svr.Close()

	t.Run("enqueue task", func(t *testing.T) {
		keeper := New(context.Background(), svr.URL, nil)
		cb := func(a *TopupAction) error {
			// do something with action
			return nil
		}
		err := keeper.Watch("batch1", correctBatchId, keeper.url, "1", "2", "45s", cb)
		if err != nil {
			t.Fatal(err)
		}

		tasks := keeper.List()
		v := tasks[0].(map[string]interface{})
		if v["active"] != true {
			t.Fatalf("there should not be any tasks in the worker")
		}
		keeper.Stop()
	})

	t.Run("dequeue task", func(t *testing.T) {
		keeper := New(context.Background(), svr.URL, nil)
		cb := func(a *TopupAction) error {
			// do something with action
			return nil
		}
		err := keeper.Watch("batch1", correctBatchId, keeper.url, "1", "2", "2s", cb)
		if err != nil {
			t.Fatal(err)
		}

		err = keeper.Unwatch(correctBatchId)
		if err != nil {
			t.Fatal(err)
		}

		tasks := keeper.List()
		v := tasks[0].(map[string]interface{})
		if v["active"] != false {
			t.Fatalf("there should not be any tasks in the worker")
		}
		keeper.Stop()
	})

	t.Run("task actions", func(t *testing.T) {
		keeper := New(context.Background(), svr.URL, nil)
		actions := []*TopupAction{}
		var mtx sync.Mutex
		cb := func(a *TopupAction) error {
			// do something with action
			mtx.Lock()
			defer mtx.Unlock()
			actions = append(actions, a)
			return nil
		}
		err := keeper.Watch("batch1", correctBatchId, keeper.url, "10000", "10000000", "10s", cb)
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Second * 2)
		info, err := keeper.GetTaskInfo(correctBatchId)
		if err != nil {
			t.Fatal(err)
		}
		if info["batch"] != correctBatchId {
			t.Fatal("batchId mismatch")
		}

		if actions[0].Name != "topup" {
			t.Fatal("first TopupAction should be topup")
		}
		if actions[1].Name != "dilute" {
			t.Fatal("second TopupAction should be dilute")
		}
		keeper.Stop()
	})
}
