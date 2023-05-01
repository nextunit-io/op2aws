package cmd

import (
	"fmt"
	"nextunit/op2aws/awsvault"
	"nextunit/op2aws/opaws"
	"os"
	"syscall"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func getNameList[T awsvault.OpInterface](list []T) []string {
	var nameList []string = []string{}

	for _, v := range list {
		nameList = append(nameList, v.GetName())
	}

	return nameList
}

func runAwsConfigCommand(configPath string) {
	if !term.IsTerminal(int(syscall.Stdin)) {
		handleError(fmt.Errorf("This functionality is not available inside of a non interactive terminal"))
	}

	commandClient := &awsvault.CommandClientDefault{}

	var profileName string
	var vaultName string
	var itemName string
	var awsAccessKeyFieldDefault = awsvault.AWS_ACCESS_KEY_FIELD_DEFAULT
	var awsSecretAccessKeyFieldDefault = awsvault.AWS_SECRET_ACCESS_KEY_FIELD_DEFAULT
	var assumeRole string
	var mfa string

	vaultList, err := awsvault.GetVaults(commandClient)
	handleError(err)

	survey.AskOne(&survey.Input{
		Message: "Enter new profile name:",
	}, &profileName, survey.WithValidator(survey.MinLength(1)))

	survey.AskOne(&survey.Select{
		Message: "Select credentials vault:",
		Options: getNameList(vaultList),
	}, &vaultName, survey.WithValidator(survey.Required))

	itemList, err := awsvault.GetItems(commandClient, vaultName)
	handleError(err)

	survey.AskOne(&survey.Select{
		Message: "Select credentials item:",
		Options: getNameList(itemList),
	}, &itemName, survey.WithValidator(survey.Required))

	assumeRoleRequired := false
	survey.AskOne(&survey.Confirm{
		Message: "Do you like to assume a specific role?",
	}, &assumeRoleRequired)

	if assumeRoleRequired {
		survey.AskOne(&survey.Input{
			Message: "Enter the role arn you'd like to assume:",
		}, &assumeRole, survey.WithValidator(survey.MinLength(20)))
	}

	mfaRequired := false
	survey.AskOne(&survey.Confirm{
		Message: "Do you like to configure MFA?",
	}, &mfaRequired)

	if mfaRequired {
		survey.AskOne(&survey.Input{
			Message: "Enter the MFA arn you'd like to assume:",
		}, &mfa, survey.WithValidator(survey.MinLength(9)))
	}

	changeDefaultLabelNames := false
	survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Do you like to change the default label names for the AWS credentials in 1password? (%s, %s)", awsAccessKeyFieldDefault, awsSecretAccessKeyFieldDefault),
	}, &changeDefaultLabelNames)

	if changeDefaultLabelNames {
		entries, err := awsvault.GetEntries(commandClient, vaultName, itemName)
		handleError(err)

		survey.AskOne(&survey.Select{
			Message: "Select the label name for the AWS_ACCESS_KEY_ID:",
			Options: getNameList(entries),
		}, &awsAccessKeyFieldDefault, survey.WithValidator(survey.Required))

		survey.AskOne(&survey.Select{
			Message: "Select the label name for the AWS_SECRET_ACCESS_KEY:",
			Options: getNameList(entries),
		}, &awsSecretAccessKeyFieldDefault, survey.WithValidator(survey.Required))
	}

	c := opaws.NewAwsConfig(&opaws.AwsConfigClientDefault{}, opaws.AWS_FILE_PATH)
	body := opaws.GetProfileBody(
		profileName,
		vaultName,
		itemName,
		assumeRole,
		mfa,
		awsAccessKeyFieldDefault,
		awsSecretAccessKeyFieldDefault,
	)

	writeFile := false
	survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf(
			"Do you like to add to the config:%s\n\nWrite now to %s?",
			body,
			c.GetPath(),
		),
	}, &writeFile)
	if writeFile {
		c.WriteProfile(body)
		fmt.Println("Added to config file.")
	}
}

func addAwsConfigCmd() {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Functionality to administrate the .aws/config file",
		Long:  "TODO",
		Run: func(cmd *cobra.Command, args []string) {
			runAwsConfigCommand(fmt.Sprintf("%s/.aws/config", os.Getenv("HOME")))
		},
	}
	rootCMD.AddCommand(cmd)
}
