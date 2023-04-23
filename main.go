package main

import (
	"fmt"
	"nextunit/op2aws/cache"
	"nextunit/op2aws/onepassword"
	"nextunit/op2aws/opaws"
	"os"
)

func handleError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	// TODO: move somewhere to configure it
	cachePath := fmt.Sprintf("%s/.aws/op2aws-cache", os.Getenv("HOME"))

	opClient := onepassword.New("nextunit.io", "AWS nextunit - Zero")
	if !opClient.CLIAvailable() {
		handleError(fmt.Errorf("1password CLI not installed. See https://developer.1password.com/docs/cli"))
	}

	cacheClient := cache.New(cachePath)
	cacheClient.GenerateFromOP(opClient)

	awsClient := opaws.New(opClient)
	awsClient.AssumeRole("arn:aws:iam::356070915882:role/Administrator")
	awsClient.UseMFA("arn:aws:iam::116675466806:mfa/zero")

	credentials, err := cacheClient.GetCache()
	handleError(err)

	if credentials == nil {
		credentials, err := awsClient.GetCredentials()
		handleError(err)

		err = cacheClient.Store(credentials)
		handleError((err))

		fmt.Println(credentials)
	} else {
		fmt.Println(credentials)
	}
}
