/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"evo-cli/internal"
	"fmt"

	"github.com/spf13/cobra"
)

// inspireCmd represents the inspire command
var inspireCmd = &cobra.Command{
	Use:   "inspire",
	Short: "Get inspired",
	Long:  `Get inspired by a random quote from a list of quotes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(internal.GetQuote())
	},
}

func init() {
	rootCmd.AddCommand(inspireCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inspireCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inspireCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
