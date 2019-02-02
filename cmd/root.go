package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "s3cret",
	Short: "s3cret copies files to/from AWS S3 with automatic client-side encryption/decryption",
	Long: `s3cret copies files to/from AWS S3 with automatic client-side encryption/decryption
See https://github.com/luhring/s3cret for documentation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Test!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
