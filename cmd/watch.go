/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	uds "github.com/asabya/go-ipc-uds"
	"github.com/spf13/cobra"
)

var (

	// watchCmd represents the watch command
	watchCmd = &cobra.Command{
		Use:   "watch",
		Short: "Watch a batch",
		Long: `Watch will start a worker and keep watching the stamp
based on the provided parameters`,
		Run: func(cmd *cobra.Command, args []string) {
			if !uds.IsIPCListening(socketPath) {
				cmd.Println("Please start the keeper to run this command")
				return
			}
			batchId, err := cmd.Flags().GetString("batch")
			if err != nil {
				cmd.Printf("Failed to read batch flag %s\n", err.Error())
				return
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				cmd.Printf("Failed to read name flag %s\n", err.Error())
				return
			}
			url, err := cmd.Flags().GetString("url")
			if err != nil {
				cmd.Printf("Failed to read url flag %s\n", err.Error())
				return
			}
			interval, err := cmd.Flags().GetString("interval")
			if err != nil {
				cmd.Printf("Failed to read interval flag %s\n", err.Error())
				return
			}
			minBalance, err := cmd.Flags().GetString("min")
			if err != nil {
				cmd.Printf("Failed to read min flag %s\n", err.Error())
				return
			}
			topupAmount, err := cmd.Flags().GetString("top")
			if err != nil {
				cmd.Printf("Failed to read top flag %s\n", err.Error())
				return
			}

			if batchId == "" {
				cmd.Println("Please provide a valid batch id")
				return
			}
			if name == "" {
				cmd.Println("Please provide an unique name")
				return
			}
			if interval == "" {
				cmd.Println("Please provide a time interval to check balance")
				return
			}
			if url == "" {
				cmd.Println("Please provide the url to check balance")
				return
			}
			if minBalance == "" {
				cmd.Println("Please provide the min balance for the batch")
				return
			}
			if topupAmount == "" {
				cmd.Println("Please provide the amount to be topped up")
				return
			}
			if handler == nil {
				cmd.Println("Please run start command before watch")
				return
			}

			if err := handler.Watch(name, batchId, url, minBalance, topupAmount, interval); err != nil {
				cmd.Printf("Failed to watch %s: %s\n", batchId, err.Error())
				return
			}

			cmd.Printf("Successfully started stampkeeping on %s\n", batchId)
		},
	}
)

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().String("name", "", "Custom identifier")
	watchCmd.Flags().String("batch", "", "BatchId to topup")
	watchCmd.Flags().String("interval", "30s", "Interval to check for balance")
	watchCmd.Flags().String("url", "", "Endpoint to check balance")
	watchCmd.Flags().String("min", "2000000", "Minimum balance for topup")
	watchCmd.Flags().String("top", "5000000", "Amount to be topped up")
}
