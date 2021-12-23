/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fairDataSociety/stampkeeper/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	server string

	// startCmd represents the start command
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start stampkeeper",
		Long:  `Start the stampkeeper to run in the background`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if server == "" {
				return fmt.Errorf("server endpoind is missing. please run \"--help\"")
			}

			home, err := os.UserHomeDir()
			if err != nil {
				cmd.Println("Failed to get home location")
				return err
			}
			cb := func(a *pkg.TopupAction) error {
				// do something with action
				logger.Infof("Got action %+v", a)
				f, err := os.OpenFile(filepath.Join(home, accountant), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					return err
				}

				defer f.Close()
				b, err := json.Marshal(a)
				if err != nil {
					return err
				}

				if _, err = f.WriteString(fmt.Sprintf("%s\n", b)); err != nil {
					return err
				}
				return nil
			}
			keeper = pkg.New(ctx, server, logger)
			batches := viper.Get("batches")
			b := batches.(map[string]interface{})
			for i, v := range b {
				value := v.(map[string]interface{})
				if value["active"] == "true" {
					err := keeper.Watch(
						value["name"].(string),
						i,
						value["url"].(string),
						value["min"].(string),
						value["top"].(string),
						value["interval"].(string),
						cb,
					)
					if err != nil {
						cmd.Println(err)
						return err
					}
				}
			}
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)

			select {
			case <-ctx.Done():
			case <-c:
				cancel()
			}
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&server, "server", "", "dfs server api endpoint")
}
