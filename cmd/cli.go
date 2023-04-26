package cmd

import (
	"encoding/json"
	"fmt"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/cache"
	"nextunit/op2aws/config"
	"nextunit/op2aws/opaws"
	"os"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

func runAwsCliCommand(vault, item, mfaArn, assumeRoleArn string, forceCache bool, export bool, awsAccessKeyFieldDefault, awsSecretAccessKeyFieldDefault string) {
	opClient := awsvault.NewOnePasswordVault(vault, item)
	opClient.SetDefaults(awsAccessKeyFieldDefault, awsSecretAccessKeyFieldDefault, "TODO")

	awsClient := opaws.New(opClient)
	awsClient.UseMFA(mfaArn)
	awsClient.AssumeRole(assumeRoleArn)

	cacheClient := cache.New(os.Getenv("HOME"))
	cacheClient.GenerateFromOP(opClient)
	cacheClient.GenerateFromOPAWS(awsClient)

	var credentials *sts.Credentials
	cacheCredentials, err := cacheClient.GetCache()
	handleError(err)

	if cacheCredentials == nil || forceCache {
		for i := 0; i < 1; i++ {
			c, err := awsClient.GetCredentials()
			handleError(err)

			if c != nil {
				credentials = c
				break
			}
		}

		err = cacheClient.Store(credentials)
		handleError((err))
	} else {
		credentials = cacheCredentials
	}

	if export {
		fmt.Println("export AWS_ACCESS_KEY_ID=" + *credentials.AccessKeyId)
		fmt.Println("export AWS_SECRET_ACCESS_KEY=" + *credentials.SecretAccessKey)
		fmt.Println("export AWS_SESSION_TOKEN=" + *credentials.SessionToken)
	} else {
		credentialsByteString, err := json.Marshal(credentials)
		handleError(err)

		var credentialsInterface map[string]interface{}
		err = json.Unmarshal(credentialsByteString, &credentialsInterface)
		handleError(err)

		credentialsInterface["Version"] = 1
		output, err := json.Marshal(credentialsInterface)
		handleError(err)

		fmt.Println(string(output))
	}
}

func addAwsCliCmd() {
	var mfaArn string
	var assumeRoleArn string
	var forceCache bool
	var export bool
	var awsAccessKeyFieldDefault string
	var awsSecretAccessKeyFieldDefault string

	cmd := &cobra.Command{
		Use:   config.COMMAND_CLI,
		Short: "Functionality to use inside of the .aws/config file",
		Long:  "This function can be used inside of the .aws/config file as profile:\n\n[profile nextunit]\n   credential_process = sh -c '\"" + config.COMMAND_ROOT + "\" \"cli-profile\" \"1password-vault\" \"1password-item\" \"-m\" \"mfa-arn\" \"-a\" \"assume-role-arn\"'",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			runAwsCliCommand(args[0], args[1], mfaArn, assumeRoleArn, forceCache, export, awsAccessKeyFieldDefault, awsSecretAccessKeyFieldDefault)
		},
	}
	cmd.Flags().StringVarP(&mfaArn, "mfa", "m", "", "When using 1password MFA it is possible to use this flag to specify the MFA arn")
	cmd.Flags().StringVarP(&assumeRoleArn, "assume-role", "a", "", "To assume a specific role when getting the credentials, it is possible to use this flat for adding the arn of the role")
	cmd.Flags().BoolVarP(&forceCache, "force", "f", false, "To force the execution without using the cache")
	cmd.Flags().BoolVarP(&export, "export", "e", false, "To get the export command. It can be used to run it via `source $(op2aws cli ... --export)`")
	cmd.Flags().StringVarP(&awsAccessKeyFieldDefault, "label-accesskey", "k", awsvault.AWS_ACCESS_KEY_FIELD_DEFAULT, "To override the label field name in 1password for the AWS_ACCESS_KEY_ID")
	cmd.Flags().StringVarP(&awsSecretAccessKeyFieldDefault, "label-secret-accesskey", "s", awsvault.AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT, "To override the label field name in 1password for the AWS_SECRET_ACCESS_KEY")
	rootCMD.AddCommand(cmd)
}
