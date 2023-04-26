package cmd

import (
	"fmt"
	"nextunit/op2aws/config"
	"os"

	"github.com/spf13/cobra"
)

var rootCMD *cobra.Command

func init() {
	rootCMD = &cobra.Command{
		Use:   config.COMMAND_ROOT + " [command]",
		Short: config.COMMAND_ROOT + " is a tool to provide AwS credentials from 1password",
		Long:  config.COMMAND_ROOT + " is a tool that provides AWS credentials for the CLI or other AWS SDKs/CLIs by using 1password and generate credentials for MFA / Assume Roles in AWS or just by storing the hardcoded credentials in 1password",
	}

	addAwsCliCmd()
	addAwsConfigCmd()
}

func Execute() {
	rootCMD.Execute()
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
