/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	uds "github.com/asabya/go-ipc-uds"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop stampkeeper",
	Long:  `Stops The stampkeeper and all the watchers`,
	Run: func(cmd *cobra.Command, args []string) {
		if !uds.IsIPCListening(socketPath) {
			cmd.Println("Please start the keeper to run this command")
			return
		}
		if handler == nil {
			cmd.Println("Please start the keeper to run this command")
			return
		}
		if cancel != nil {
			cancel()
		}
		cmd.Println("Stopped stampkeeper")
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
