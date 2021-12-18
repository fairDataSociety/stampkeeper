package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	correctBatchId = "6a50032864056992563cee7e31b3323bd25ac34c157f658d02b32a59e241f7f3"
	initialAmount  = "1234"
)

type mockResponse struct {
	BatchID string `json:"batchID"`
}

func TestTopupTask(t *testing.T) {
	stampInfo := &stamp{
		BatchID: correctBatchId,
		Amount:  initialAmount,
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

		_, err := newTopupTask(context.Background(), "batch1", wbi, svr.URL, svr.URL, minAmount, topAmount, time.Second*10)
		if err == nil {
			t.Fatal("wrong batch id check failed")
		}
	})

	t.Run("correct batch id", func(t *testing.T) {
		minAmount := &big.Int{}
		minAmount.SetString("10000", 10)

		topAmount := &big.Int{}
		topAmount.SetString("10000000", 10)

		topupTask, err := newTopupTask(context.Background(), "batch1", correctBatchId, svr.URL, svr.URL, minAmount, topAmount, time.Second*10)
		if err != nil {
			t.Fatal(err)
		}
		if topupTask.Name() != "batch1" {
			t.Fatal("task name mismatch")
		}

		go func() {
			err = topupTask.Execute(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
		}()
		// wait for first run
		<-time.After(time.Second * 3)
		topupTask.Stop()
		if stampInfo.Amount != "10001234" {
			t.Fatal("topup failed")
		}
	})
}

// test for same name
// test with multiple batchIds
