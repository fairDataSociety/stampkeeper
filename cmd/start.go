/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fairDataSociety/stampkeeper/pkg"
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
			keeper = pkg.New(ctx, server)

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
