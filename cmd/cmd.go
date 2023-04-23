package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCMD *cobra.Command

func init() {
	rootCMD = &cobra.Command{
		Use:   "op2aws [command]",
		Short: "op2aws is a tool to provide AwS credentials from 1password",
		Long:  "op2aws is a tool that provides AWS credentials for the CLI or other AWS SDKs/CLIs by using 1password and generate credentials for MFA / Assume Roles in AWS or just by storing the hardcoded credentials in 1password",
	}

	addAwsProfileCLICMD()
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
