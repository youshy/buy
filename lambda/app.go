package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.uber.org/zap"
)

type config struct {
	Page            string
	Product         string
	SoldOutString   string
	AddToCardString string
	SendEmailTo     string
	Email           bool
	SMS             bool
	SendSMSTo       string
	Session         *session.Session
	Log             *zap.SugaredLogger
}

func newConfig() config {
	c := config{
		Page:            os.Getenv("PAGE"),
		Product:         os.Getenv("PRODUCT"),
		SoldOutString:   os.Getenv("SOLD_OUT_STRING"),
		AddToCardString: os.Getenv("ADD_TO_CARD_STRING"),
		SendEmailTo:     os.Getenv("SEND_EMAIL_TO"),
	}

	email, err := strconv.ParseBool(os.Getenv("EMAIL"))
	if err != nil {
		c.Log.Fatalf("Unable to parse EMAIL %v", err)
	}
	c.Email = email

	sms, err := strconv.ParseBool(os.Getenv("SMS"))
	if err != nil {
		c.Log.Fatalf("Unable to parse SMS %v", err)
	}
	c.SMS = sms

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushing buffer?
	c.Log = logger.Sugar()

	c.Session = session.New()

	if c.Page == "" {
		c.Log.Fatal("Page not set")
	}
	if c.Product == "" {
		c.Log.Fatal("Product not set")
	}
	if c.SoldOutString == "" {
		c.Log.Fatal("SoldOutString not set")
	}
	if c.AddToCardString == "" {
		c.Log.Fatal("AddToCardString not set")
	}
	if c.Email {
		if c.SendEmailTo == "" {
			c.Log.Fatal("SendEmailTo not set")
		}
	}
	if c.SMS {
		if c.SendSMSTo == "" {
			c.Log.Fatal("SendSMSTo not set")
		}
	}
	return c
}

func (c *config) EmailSubjectString() string {
	return fmt.Sprintf("%s is available again!", c.Product)
}

func (c *config) BodyString() string {
	return fmt.Sprintf("%s is available for purchase on %s. This was generated on %v\n\nGenerated by automated AWS Lambda service.", c.Product, c.Page, time.Now())
}

func handleRequest() error {
	c := newConfig()

	// Fetch page
	res, err := http.Get(c.Page)
	if err != nil {
		c.Log.Errorf("Unable to fetch page: %v", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		c.Log.Errorf("Page fetch status: %d", res.StatusCode)
		return err
	}

	// Check if the string contains the "SoldOut" string
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		c.Log.Errorf("Unable to build the doc from Body: %v", err)
		return err
	}

	var (
		spanRes   []string
		soldOut   int
		buttonRes []string
		emailSent bool
		smsSent   bool
	)

	// find SoldOut
	doc.Find("span").Each(func(i int, s *goquery.Selection) {
		spanRes = append(spanRes, s.Text())
	})

	for _, v := range spanRes {
		if v == c.SoldOutString {
			soldOut += 1
		}
	}

	// find AddToCard
	doc.Find("button").Each(func(i int, s *goquery.Selection) {
		buttonRes = append(buttonRes, s.Text())
	})

	if soldOut == 0 && len(buttonRes) > 0 {
		if c.Email {
			ok := c.sendEmail()
			if !ok {
				return errors.New("Unable to send the email")
			}
			emailSent = true
		}
		if c.SMS {
			ok := c.sendSMS()
			if !ok {
				return errors.New("Unable to send the sms")
			}
			smsSent = true
		}
	}

	c.Log.Infof("Checked the %s. Email [Enabled: %v]: %v SMS [Enabled: %v]: %v\n(true - product is available, false - product is unavailable", c.Product, c.Email, emailSent, c.SMS, smsSent)
	return nil
}