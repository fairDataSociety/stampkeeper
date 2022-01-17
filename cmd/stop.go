/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
		defer func() {
			if cancel != nil {
				cancel()
			}
		}()
		home, err := os.UserHomeDir()
		if err != nil {
			cmd.Println("Failed to get home location")
			return
		}
		filename := filepath.Join(home, fmt.Sprintf("stampkeeper_history_%d.json", time.Now().Unix()))
		f, err := os.Create(filename)
		if err != nil {
			cmd.Println("Failed to create funding history")
			return
		}
		defer f.Close()
		list := keeper.List()
		b, err := json.MarshalIndent(list, "", "\t")
		if err != nil {
			cmd.Println("Failed to read batch list")
			return
		}

		_, err = f.Write(b)
		if err != nil {
			cmd.Println("Failed to create funding history")
			return
		}

		cmd.Println("Stopped stampkeeper. Topup history saved in ", filename)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
