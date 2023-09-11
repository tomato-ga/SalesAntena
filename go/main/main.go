package main

import (
	"Salesscrape/headless"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedResult struct {
	Feed *gofeed.Feed
	Err error
	Tags []string

}

func FeedScrape(url string) {
	customClinet := &http.Client{
		Timeout: 4* time.Second,
	}

	fp := gofeed.NewParser()
	fp.Client = customClinet
	feed, err := fp.ParseURL(url)
	if err != nil {
		return
	}

	for i, item := range feed.Items {
		println(i, item.Title, item.Link)
	}

}

func main() {

	topUrl := headless.TimesalePage()
	fmt.Println(topUrl)
	fmt.Println(len(topUrl))

}

