package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"

	"github.com/luhring/s3cret/s3cret"
)

func init() {
	rootCmd.AddCommand(cpCmd)
}

var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "Copy file to or from an S3 bucket, using client-side encryption/decryption",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// For now, we're assuming the first argument (source) is local and the second argument (destination) is S3

		localPath := args[0]
		s3URL := args[1]

		s3Object, err := s3cret.DeserializeS3URL(s3URL)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to process S3 URL: %v\n", err.Error())
			return
		}

		awsSession := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))

		client := s3cret.NewClient(awsSession)
		err = client.SendToS3(localPath, s3Object.Bucket, s3Object.Key)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable send object to S3: %v\n", err.Error())
			return
		}
	},
}
