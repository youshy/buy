package main

import (
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

type config struct {
	Page            string
	Product         string
	SoldOutString   string
	AddToCardString string
	Log             *zap.SugaredLogger
}

func newConfig() config {
	c := config{
		Page:            os.Getenv("PAGE"),
		Product:         os.Getenv("PRODUCT"),
		SoldOutString:   os.Getenv("SOLD_OUT_STRING"),
		AddToCardString: os.Getenv("ADD_TO_CARD_STRING"),
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushing buffer?
	c.Log = logger.Sugar()

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
	return c
}

func main() {
	runtime.Start(handleRequest)
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
		c.sendEmail()
	}

	return nil
}

func (c *config) sendEmail() {

}
