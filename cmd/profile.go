package cmd

import (
	"fmt"
	"nextunit/op2aws/cache"
	"nextunit/op2aws/onepassword"
	"nextunit/op2aws/opaws"
	"os"

	"github.com/spf13/cobra"
)

func runAwsProfileCliCommand(vault, item, mfaArn, assumeRoleArn string, forceCache bool) {
	opClient := onepassword.New(vault, item)

	awsClient := opaws.New(opClient)
	awsClient.UseMFA(mfaArn)
	awsClient.AssumeRole(assumeRoleArn)

	cacheClient := cache.New(os.Getenv("HOME"))
	cacheClient.GenerateFromOP(opClient)
	cacheClient.GenerateFromOPAWS(awsClient)

	credentials, err := cacheClient.GetCache()
	handleError(err)

	if credentials == nil || forceCache {
		credentials, err := awsClient.GetCredentials()
		handleError(err)

		err = cacheClient.Store(credentials)
		handleError((err))

		fmt.Println(credentials)
	} else {
		fmt.Println(credentials)
	}
}

func addAwsProfileCLICMD() {
	var mfaArn string
	var assumeRoleArn string
	var forceCache bool

	cmd := &cobra.Command{
		Use:   "cli-profile",
		Short: "Functionality to use inside of the .aws/config file",
		Long:  "This function can be used inside of the .aws/config file as profile:\n\n[profile nextunit]\n   credential_process = sh -c '\"op2aws\" \"cli-profile\" \"1password-vault\" \"1password-item\" \"-m\" \"mfa-arn\" \"-a\" \"assume-role-arn\"",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runAwsProfileCliCommand(args[0], args[1], mfaArn, assumeRoleArn, forceCache)
		},
	}
	cmd.Flags().StringVarP(&mfaArn, "mfa", "m", "", "When using 1password MFA it is possible to use this flag to specify the MFA arn")
	cmd.Flags().StringVarP(&assumeRoleArn, "assume-role", "a", "", "To assume a specific role when getting the credentials, it is possible to use this flat for adding the arn of the role")
	cmd.Flags().BoolVarP(&forceCache, "force", "f", false, "To force the execution without using the cache")
	rootCMD.AddCommand(cmd)
}
