/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	uds "github.com/asabya/go-ipc-uds"

	"github.com/spf13/cobra"
)

// unwatchCmd represents the unwatch command
var unwatchCmd = &cobra.Command{
	Use:   "unwatch",
	Short: "Stop watching a batch",
	Long:  `Unwatch will stop the watching the provided batch`,
	Run: func(cmd *cobra.Command, args []string) {
		if !uds.IsIPCListening(socketPath) {
			cmd.Println("Please start the keeper to run this command")
			return
		}
		if handler == nil {
			cmd.Println("Please run start command before unwatch")
			return
		}
		batchId, err := cmd.Flags().GetString("batch")
		if err != nil {
			cmd.Printf("Failed to read batch flag %s\n", err.Error())
			return
		}
		if batchId == "" {
			cmd.Println("Please provide a valid batch id")
			return
		}
		if err := handler.Unwatch(batchId); err != nil {
			cmd.Printf("Failed to watch %s: %s\n", batchId, err.Error())
			return
		}
		cmd.Printf("Successfully stopped stampkeeping on %s\n", batchId)
	},
}

func init() {
	rootCmd.AddCommand(unwatchCmd)

	unwatchCmd.Flags().String("batch", "", "BatchId to unwatch")
}
