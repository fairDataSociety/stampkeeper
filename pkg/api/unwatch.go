package api

import (
	"fmt"

	"github.com/spf13/viper"
)

func (h *Handler) Unwatch(batchId string) error {
	if err := h.stampkeeper.Unwatch(batchId); err != nil {
		return err
	}
	b := viper.Get(fmt.Sprintf("batches.%s", batchId))
	a := b.(map[string]interface{})
	a["active"] = "false"
	viper.Set(fmt.Sprintf("batches.%s", batchId), a)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}
