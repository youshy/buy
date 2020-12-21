package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

func (c *config) sendSMS() bool {
	svc := sns.New(c.Session)
	params := &sns.PublishInput{
		Message:     aws.String(c.BodyString()),
		PhoneNumber: aws.String(c.SendSMSTo),
	}

	result, err := svc.Publish(params)
	if err != nil {
		c.Log.Fatalf("Error whilst sending SMS: %v", err)
		return false
	}

	c.Log.Infof("Return from sms send: %v", result)
	return true
}
