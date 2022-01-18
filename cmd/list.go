/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	"encoding/json"

	uds "github.com/asabya/go-ipc-uds"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List watched stamps this session",
	Long: `List command will list all the stamps watched this session
along will some stats`,
	Run: func(cmd *cobra.Command, args []string) {
		if !uds.IsIPCListening(socketPath) {
			cmd.Println("Please start the keeper to run this command")
			return
		}
		if handler == nil {
			cmd.Println("Please run start command before list")
			return
		}
		list := handler.List()
		b, err := json.MarshalIndent(list, "", "\t")
		if err != nil {
			cmd.Println("Failed to read batch list")
			return
		}
		cmd.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// TODO list all, active, inactive
	// TODO Print Stats : last run, how many times toppedup, next run
}
