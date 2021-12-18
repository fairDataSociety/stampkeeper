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
)

func TestTaskManager(t *testing.T) {
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

	t.Run("enqueue task", func(t *testing.T) {
		keeper := New(context.Background(), svr.URL)
		err := keeper.Watch("batch1", correctBatchId, keeper.url, "1", "2", "45s")
		if err != nil {
			t.Fatal(err)
		}

		tasks := keeper.List()
		if tasks[0] != correctBatchId {
			t.Fatalf("batchId Mismatch. Got %s instead of %s ", tasks[0], correctBatchId)
		}
		keeper.Stop()
	})

	t.Run("dequeue task", func(t *testing.T) {
		keeper := New(context.Background(), svr.URL)
		err := keeper.Watch("batch1", correctBatchId, keeper.url, "1", "2", "45s")
		if err != nil {
			t.Fatal(err)
		}

		err = keeper.Unwatch(correctBatchId)
		if err != nil {
			t.Fatal(err)
		}

		tasks := keeper.List()
		if len(tasks) != 0 {
			t.Fatalf("there should not be any tasks in the worker")
		}
		keeper.Stop()
	})
}
