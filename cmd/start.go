/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start stampkeeper",
	Long:  `Start the stampkeeper to run in the background`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO add url flag, url defaults to local bee debug api
		// Start new Keeper and wait for signal for stop
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
}
