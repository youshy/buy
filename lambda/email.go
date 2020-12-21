package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ses"
)

func (c *config) sendEmail() bool {
	svc := ses.New(c.Session)

	result, err := svc.SendEmail(c.buildEmail())
	if err != nil {
		c.handleEmailErrors(err)
		return false
	}

	c.Log.Infof("Return from email send: %v", result)
	return true
}

func (c *config) buildEmail() *ses.SendEmailInput {
	return &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(c.SendEmailTo),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(c.EmailSubjectString()),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(c.BodyString()),
				},
			},
		},
		ReturnPath:    aws.String(""),
		ReturnPathArn: aws.String(""),
		Source:        aws.String("sender@example.com"), // TODO: Change that to AWS provided env var
		SourceArn:     aws.String(""),
	}
}

func (c *config) handleEmailErrors(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case ses.ErrCodeMessageRejected:
			c.Log.Fatalf("Error: %v\nDetails: %v", ses.ErrCodeMessageRejected, aerr.Error())
		case ses.ErrCodeMailFromDomainNotVerifiedException:
			c.Log.Fatalf("Error: %v\nDetails: %v", ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
		case ses.ErrCodeConfigurationSetDoesNotExistException:
			c.Log.Fatalf("Error: %v\nDetails: %v", ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
		case ses.ErrCodeConfigurationSetSendingPausedException:
			c.Log.Fatalf("Error: %v\nDetails: %v", ses.ErrCodeConfigurationSetSendingPausedException, aerr.Error())
		case ses.ErrCodeAccountSendingPausedException:
			c.Log.Fatalf("Error: %v\nDetails: %v", ses.ErrCodeAccountSendingPausedException, aerr.Error())
		default:
			c.Log.Fatalf("Error: %v", aerr.Error())
		}
	} else {
		c.Log.Fatal(err)
	}
}
