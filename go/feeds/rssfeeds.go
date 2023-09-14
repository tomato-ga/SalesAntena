package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// TODO h1, h2, h3を取得する

func extractAmazonLinks(url string) (string, []string) {
	var content string
	var amazonLinks []string

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching URL:", err)
		return content, amazonLinks
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Error reading the document:", err)
		return content, amazonLinks
	}

	// Extract main content (This depends on the website's structure. Here's a generic example.)
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		content += s.Text() + "\n"
	})

	// Extract amazon links
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		if strings.Contains(link, "amazon.co.jp") || strings.Contains(link, "amzn") {
			amazonLinks = append(amazonLinks, link)
		}
	})

	return content, amazonLinks
}

func main() {
	fp := gofeed.NewParser()

	for _, feedURL := range RssLists {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
			fmt.Println("URL:", item.Link)
			fmt.Println("Title:", item.Title)

			content, amazonLinks := extractAmazonLinks(item.Link)

			fmt.Println("Content:", content)
			fmt.Println("Amazon Links:", amazonLinks)
			fmt.Println("-------------------------")
		}
	}
}