package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// TODO h1, h2, h3を取得する

type AmazonKeyValuePair struct {
	ASIN     string
	URL      string
	ImageURL string
}

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

func transformAmazonID(urls []string, imageLinks []string) []AmazonKeyValuePair {
	regexForASIN := regexp.MustCompile(`[A-Za-z0-9]{10}`)
	regexForTag := regexp.MustCompile(`\w+-22`)

	var results []AmazonKeyValuePair

	// ASINを重複しないようにするためのマップ
	seenASIN := make(map[string]bool)

	for index, url := range urls {
		ASIN := regexForASIN.FindString(url)
		TAG := regexForTag.FindString(url)
		newURL := strings.Replace(url, TAG, "entamenews-22", 1)

		var imageURL string
		if index < len(imageLinks) {
			imageURL = imageLinks[index]
		}

		if !seenASIN[ASIN] {
			results = append(results, AmazonKeyValuePair{
				ASIN:     ASIN,
				URL:      newURL,
				ImageURL: imageURL,
			})
			seenASIN[ASIN] = true
		}
	}
	return results
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

func extractAmazonLinks(url string, config FeedConfig) (string, string, string, []string, []string) {
	var content string
	var amazonLinks []string
	var amazonImageLinks []string
	var h1 string
	var h2 string

	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error fetching URL:", err)
		return h1, h2, content, amazonLinks, amazonImageLinks
	}
	log.Println("Content-Type:", resp.Header.Get("Content-Type"))
	defer resp.Body.Close()

	var reader io.Reader
	if resp.Header.Get("Content-Type") == "text/html; charset=euc-jp" {
		reader = transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder())
	} else {
		reader = resp.Body
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Println("Error reading the document:", err)
		return h1, h2, content, amazonLinks, amazonImageLinks
	}

	doc.Find(config.Selector).Eq(0).Each(func(i int, s *goquery.Selection) {
		content = s.Text()
		for _, removePhrase := range config.RemoveText {
			content = strings.ReplaceAll(content, removePhrase, "")
			content = strings.TrimSpace(content)
			re := regexp.MustCompile(`\s+`)
			content = re.ReplaceAllString(content, " ")
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

			switch {
			case strings.Contains(link, "amazon.co.jp"):
				amazonLinks = append(amazonLinks, link)
			case strings.Contains(link, "amzn"):
				amazonLinks = append(amazonLinks, link)
			case strings.Contains(link, "valuecommerce"):
				amazonURL, err := extractAmazonURL(link)
				if err == nil && (strings.Contains(amazonURL, "amazon.co.jp") || strings.Contains(amazonURL, "amzn")) {
					amazonLinks = append(amazonLinks, amazonURL)
				}
			}
		})

		s.Find("img").Each(func(j int, ImageLinkElement *goquery.Selection) {
			imagelink, exists := ImageLinkElement.Attr("src")
			if exists && strings.Contains(imagelink, "amazon") {
				amazonImageLinks = append(amazonImageLinks, imagelink)
			}
		})
	})

	// すべてのリンクが追加された後に、重複を削除
	amazonLinks = uniqueAmazonURL(amazonLinks)

	return h1, h2, content, amazonLinks, amazonImageLinks
}

func main() {
	fp := gofeed.NewParser()

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
			fmt.Println("URL:", item.Link)
			fmt.Println("Title:", item.Title)

			h1, h2, content, amazonLinks, amazonImageLinks := extractAmazonLinks(item.Link, config)

			if len(amazonLinks) > 0 {
				results := transformAmazonID(amazonLinks, amazonImageLinks)

				fmt.Println("H1:", h1)
				fmt.Println("H2:", h2)
				fmt.Println("Content:", content)
				fmt.Println(results)

				// for _, res := range results {
				// 	fmt.Println("ASIN:", res.ASIN)
				// 	fmt.Println("Amazon Link:", res.URL)
				// 	fmt.Println("Amazon Image:", res.ImageURL)
				// }
				fmt.Println("-------------------------")
			}
		}
	}
}
