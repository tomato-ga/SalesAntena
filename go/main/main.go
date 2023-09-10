package main

import (
	"Salesscrape/headless"
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
	// for _, url := range RssLists {
	// 	FeedScrape(url)
	// }

	headless.Heads()
}