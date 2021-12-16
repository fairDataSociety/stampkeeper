/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start stampkeeper",
	Long: `Start the stampkeeper to run in the background`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
}
