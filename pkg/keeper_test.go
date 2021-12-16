package pkg

import (
	"context"
	"testing"
	"time"
)

func TestTaskManager(t *testing.T) {
	t.Run("watch batchid", func(t *testing.T) {
		keeper := New(context.Background())
		err := keeper.Watch("someBatchId", "2s")
		if err != nil {
			t.Fatal(err)
		}
		<-time.After(time.Second * 4)
		keeper.Stop()
	})
}
