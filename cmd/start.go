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

	uds "github.com/asabya/go-ipc-uds"
	"github.com/fairDataSociety/stampkeeper/pkg/api"
	"github.com/fairDataSociety/stampkeeper/pkg/bot/telegram"
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
			if !uds.IsIPCListening(socketPath) {
				return fmt.Errorf("server already running")
			}
			handler = api.NewHandler(ctx, server, logger)

			token := viper.Get("telegram_bot_token")
			if token == nil {
				return fmt.Errorf("bot token not available")
			}
			botHandler, err := telegram.NewBot(ctx, fmt.Sprintf("%v", token), handler, logger)
			if err != nil {
				logger.Errorf("failed to create bot instance")
				return err
			}
			handler.SetBot(botHandler)
			err = handler.StartWatchingAll()
			if err != nil {
				logger.Errorf("failed to start watching all batches")
				return err
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
