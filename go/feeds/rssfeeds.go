package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

// TODO h1, h2, h3を取得する


func extractAmazonURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	vcURL := parsedURL.Query().Get("vc_url")
	if vcURL == "" {
		return "", fmt.Errorf("vc_url not Found")
	}

	decodedURL, err := url.QueryUnescape(vcURL)
	if err != nil {
		return "", err
	}
	return decodedURL, nil

}


func uniqueAmazonURL(amazonURLs []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range amazonURLs {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}


func extractAmazonLinks(url string, config FeedConfig) (string, string, string, []string) {
	var content string
	var amazonLinks []string
	var h1 string
	var h2 string

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching URL:", err)
		return h1, h2, content, amazonLinks
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println("Error reading the document:", err)
		return h1, h2, content, amazonLinks
	}


	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		content = s.Text()
		for _, removePhrase := range config.RemoveText {
			content = strings.ReplaceAll(content, removePhrase, "")
		}
	})
	

	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		h1 = s.Text()
	}) 

	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		h2 = s.Text()
	}) 

	// Extract amazon links
	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		s.Find("a[href]").Each(func(j int, linkElement *goquery.Selection) {
			link, _ := linkElement.Attr("href")
			if strings.Contains(link, "amazon.co.jp") || strings.Contains(link, "amzn") {
				amazonLinks = append(amazonLinks, link)
			} else if strings.Contains(link, "valuecommerce") {
				amazonURL, err := extractAmazonURL(link)
				if err == nil && (strings.Contains(amazonURL, "amazon.co.jp") || strings.Contains(amazonURL, "amzn")) {
					amazonLinks = append(amazonLinks, amazonURL)
				}
			}
		})
	})

	// すべてのリンクが追加された後に、重複を削除
	amazonLinks = uniqueAmazonURL(amazonLinks)

	return h1, h2, content, amazonLinks
}

func main() {
	fp := gofeed.NewParser()

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		// feed.Itemsが空でない場合、最初のアイテムのみを処理
		if len(feed.Items) > 0 {
			item := feed.Items[0]

			fmt.Println("URL:", item.Link)
			fmt.Println("Title:", item.Title)

			h1, h2, content, amazonLinks := extractAmazonLinks(item.Link, config)

			if len(amazonLinks) > 0 {
				fmt.Println("H1:", h1)
				fmt.Println("H2:", h2)
				fmt.Println("Content:", content)
				fmt.Println("Amazon Links:", amazonLinks)
				fmt.Println("-------------------------")
			}
		}
	}
}