/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	uds "github.com/asabya/go-ipc-uds"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// unwatchCmd represents the unwatch command
var unwatchCmd = &cobra.Command{
	Use:   "unwatch",
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
		batchId, err := cmd.Flags().GetString("batch")
		if err != nil {
			cmd.Printf("Failed to read batch flag %s\n", err.Error())
			return
		}
		if batchId == "" {
			cmd.Println("Please provide a valid batch id")
			return
		}
		if err := keeper.Unwatch(batchId); err != nil {
			cmd.Printf("Failed to watch %s: %s\n", batchId, err.Error())
			return
		}
		cmd.Printf("Successfully stopped stampkeeping on %s\n", batchId)
		b := viper.Get(fmt.Sprintf("batches.%s", batchId))
		a := b.(map[string]interface{})
		a["active"] = "false"
		viper.Set(fmt.Sprintf("batches.%s", batchId), a)
		if err := viper.WriteConfig(); err != nil {
			cmd.Printf("Failed to write config with batchId info %s: %s\n", batchId, err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(unwatchCmd)

	unwatchCmd.Flags().String("batch", "", "BatchId to unwatch")
}
