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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !uds.IsIPCListening(socketPath) {
			cmd.Println("Please start the keeper to run this command")
			return
		}
		if keeper == nil {
			cmd.Println("Please start the keeper to run this command")
			return
		}
		list := keeper.List()
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
