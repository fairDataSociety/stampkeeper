/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

var (

	// watchCmd represents the watch command
	watchCmd = &cobra.Command{
		Use:   "watch",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		Run: func(cmd *cobra.Command, args []string) {

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
			if keeper == nil {
				cmd.Println("Please start the keeper to run this command")
				return
			}

			if err := keeper.Watch(name, batchId, url, minBalance, topupAmount, interval); err != nil {
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
	watchCmd.Flags().String("min", "10000", "Minimum balance for topup")
	watchCmd.Flags().String("top", "1000000", "Amount to be topped up")
}
