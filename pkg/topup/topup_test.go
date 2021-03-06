package topup

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

	"github.com/fairDataSociety/stampkeeper/pkg/logging/mock"
)

var (
	correctBatchId = "6a50032864056992563cee7e31b3323bd25ac34c157f658d02b32a59e241f7f3"
	initialAmount  = "1234"
)

type mockResponse struct {
	BatchID string `json:"batchID"`
}

func TestTopupTask(t *testing.T) {
	stampInfo := &Stamp{
		BatchID:     correctBatchId,
		Amount:      initialAmount,
		Utilization: 16,
		Depth:       20,
		BucketDepth: 16,
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.String(), "/stamps/topup/") {
			amount := &big.Int{}
			amount.SetString(stampInfo.Amount, 10)
			amount = amount.Add(amount, big.NewInt(10000000))
			stampInfo.Amount = amount.String()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			_ = json.NewEncoder(w).Encode(&mockResponse{BatchID: stampInfo.BatchID})
		} else if strings.HasPrefix(r.URL.String(), "/stamps/dilute/") {
			amount := &big.Int{}
			amount.SetString(stampInfo.Amount, 10)
			amount = amount.Sub(amount, big.NewInt(5000000))
			stampInfo.Amount = amount.String()
			stampInfo.Depth = stampInfo.Depth + 2
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(202)
			_ = json.NewEncoder(w).Encode(&mockResponse{BatchID: stampInfo.BatchID})
		} else if strings.HasPrefix(r.URL.String(), "/stamps/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(stampInfo)
		} else {
			fmt.Println()
		}
	}))
	defer svr.Close()

	t.Run("topup wrong batch id", func(t *testing.T) {
		wbi := "wrongBatchId"

		minAmount := &big.Int{}
		minAmount.SetString("10000", 10)

		topAmount := &big.Int{}
		topAmount.SetString("10000000", 10)
		cb := func(a *TopupAction) error {
			// do something with action
			return nil
		}
		_, err := NewTopupTask(context.Background(), "batch1", wbi, svr.URL, svr.URL, minAmount, topAmount, time.Second*10, cb, mock.Logging{})
		if err == nil {
			t.Fatal("wrong batch id check failed")
		}
	})

	t.Run("correct batch id", func(t *testing.T) {
		minAmount := &big.Int{}
		minAmount.SetString("10000", 10)

		topAmount := &big.Int{}
		topAmount.SetString("10000000", 10)
		actions := []*TopupAction{}
		var mtx sync.Mutex
		cb := func(a *TopupAction) error {
			// do something with action
			mtx.Lock()
			defer mtx.Unlock()
			actions = append(actions, a)
			return nil
		}
		topupTask, err := NewTopupTask(context.Background(), "batch1", correctBatchId, svr.URL, svr.URL, minAmount, topAmount, time.Second*2, cb, mock.Logging{})
		if err != nil {
			t.Fatal(err)
		}
		if topupTask.Name() != "batch1" {
			t.Fatal("task Action mismatch")
		}
		go func() {
			err = topupTask.Execute(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
		}()
		// wait for first run
		<-time.After(time.Second * 5)
		topupTask.Stop()

		if stampInfo.Amount != "5001234" {
			t.Fatal("topup failed")
		}

		if actions[0].Action != "topup" {
			t.Fatal("first TopupAction should be topup")
		}
		if actions[1].Action != "dilute" {
			t.Fatal("second TopupAction should be dilute")
		}
	})
}

// TODO test for same Action
// TODO test with multiple batchIds
