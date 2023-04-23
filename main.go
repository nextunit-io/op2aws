package main

import (
	"fmt"
	"nextunit/op2aws/onepassword"
	"nextunit/op2aws/opaws"
)

func main() {
	opClient := onepassword.New("nextunit.io", "AWS nextunit - Zero")
	awsClient := opaws.New(nil, opClient)
	awsClient.AssumeRole("arn:aws:iam::356070915882:role/Administrator")
	awsClient.UseMFA("arn:aws:iam::116675466806:mfa/zero")

	fmt.Println(awsClient.GetCredentials())
}
