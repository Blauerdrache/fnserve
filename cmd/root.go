package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fnserve",
	Short: "FnServe - Lambda without the cloud",
	Long:  `FnServe is a lightweight, self-hosted function runner inspired by AWS Lambda.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
