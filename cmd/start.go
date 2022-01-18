/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fairDataSociety/stampkeeper/pkg/bot/telegram"

	uds "github.com/asabya/go-ipc-uds"
	"github.com/fairDataSociety/stampkeeper/pkg/api"

	"github.com/spf13/cobra"
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
			if uds.IsIPCListening(socketPath) {
				return fmt.Errorf("server already running")
			}
			handler = api.NewHandler(ctx, server, logger)
			botHandler, err := telegram.NewBot(handler)
			if err != nil {
				logger.Errorf("failed to create bot instance")
				return err
			}
			handler.SetBot(botHandler)

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
