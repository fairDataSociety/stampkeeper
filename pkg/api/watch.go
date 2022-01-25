package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fairDataSociety/stampkeeper/pkg/topup"
	"github.com/spf13/viper"
)

var (
	accountant = "stampkeeper_accountant.json"
)

func (h *Handler) actionCallback(a *topup.TopupAction) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	// do something with action
	f, err := os.OpenFile(filepath.Join(home, accountant), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer f.Close()
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	actionString := fmt.Sprintf("%s\n", b)
	if _, err = f.WriteString(actionString); err != nil {
		return err
	}

	message := fmt.Sprintf("Action: %s | BatchID : %s | AmountTopped: %s | DepthAdded : %d \n\n %s", a.Action, a.BatchID, a.AmountTopped, a.DepthAdded, actionString)
	if err := h.bot.Notify(message); err != nil {
		return err
	}
	return nil
}

func (h *Handler) StartWatchingAll() error {
	batches := viper.Get("batches")
	b := batches.(map[string]interface{})
	for i, v := range b {
		value := v.(map[string]interface{})
		if value["active"] == "true" {
			err := h.stampkeeper.Watch(
				value["name"].(string),
				i,
				value["url"].(string),
				value["min"].(string),
				value["top"].(string),
				value["interval"].(string),
				h.actionCallback,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *Handler) Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval string) error {
	err := h.stampkeeper.Watch(name, batchId, balanceEndpoint, minBalance, topupBalance, interval, h.actionCallback)
	if err != nil {
		return err
	}
	conf := map[string]string{}
	conf["name"] = name
	conf["interval"] = interval
	conf["url"] = balanceEndpoint
	conf["min"] = minBalance
	conf["top"] = topupBalance
	conf["active"] = "true"

	viper.Set(fmt.Sprintf("batches.%s", batchId), conf)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}
