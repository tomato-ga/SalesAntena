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

type AmazonLinkDetails struct {
	ASIN     string
	URL      string
	URLtitle string
	ImageURL string
}

type ArticleDetails struct {
	ArticleURL    string
	ArticleTitle  string
	Content       string
	AmazonDetails []AmazonLinkDetails
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

func transformAmazonID(urls []string, urlsTitle []string, imageLinks []string) []AmazonLinkDetails {
	regexForASIN := regexp.MustCompile(`[A-Za-z0-9]{10}`)
	regexForTag := regexp.MustCompile(`\w+-22`)

	var results []AmazonLinkDetails

	for index, url := range urls {
		ASIN := regexForASIN.FindString(url)
		TAG := regexForTag.FindString(url)
		newURL := strings.Replace(url, TAG, "entamenews-22", 1)

		var imageURL string
		if index < len(imageLinks) {
			imageURL = imageLinks[index]
		}

		var title string
		if index < len(urlsTitle) {
			title = urlsTitle[index]
		}

		results = append(results, AmazonLinkDetails{
			ASIN:     ASIN,
			URL:      newURL,
			URLtitle: title,
			ImageURL: imageURL,
		})
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

func extractAmazonLinks(url string, config FeedConfig) (string, []string, []string, []string) {
	var content string
	var amazonLinks []string
	var amazonLinksTitle []string
	var amazonImageLinks []string

	// User Agentの設定
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return content, amazonLinks, amazonImageLinks, amazonLinksTitle
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error fetching URL:", err)
		return content, amazonLinks, amazonImageLinks, amazonLinksTitle
	}
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
		return content, amazonLinks, amazonImageLinks, amazonLinksTitle
	}

	doc.Find(config.Selector).Eq(0).Each(func(i int, s *goquery.Selection) {
		// RemoveDivで指定された要素を削除
		for _, divToRemove := range config.RemoveDiv {
			s.Find(divToRemove).Remove()
		}

		content = s.Text()
		for _, removePhrase := range config.RemoveText {
			content = strings.ReplaceAll(content, removePhrase, "")
			content = strings.TrimSpace(content)
			re := regexp.MustCompile(`\s+`)
			content = re.ReplaceAllString(content, " ")
		}
	})

	// Extract amazon links
	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		s.Find("a[href]").Each(func(j int, linkElement *goquery.Selection) {
			link, _ := linkElement.Attr("href")

			text := linkElement.Text()
			isBlacklested := false
			for _, blackText := range BlackList["Texts"].Black {
				if text == blackText {
					isBlacklested = true
					break
				}
			}

			if isBlacklested {
				amazonLinksTitle = append(amazonLinksTitle, "リンクテキストなし")
			} else {
				amazonLinksTitle = append(amazonLinksTitle, text)
			}

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
			} else {
				amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
			}
		})
	})

	// すべてのリンクが追加された後に、重複を削除
	amazonLinks = uniqueAmazonURL(amazonLinks)

	return content, amazonLinks, amazonImageLinks, amazonLinksTitle
}

func main() {
	fp := gofeed.NewParser()

	var articles []ArticleDetails

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
			content, amazonLinks, amazonImageLinks, amazonLinksTitle := extractAmazonLinks(item.Link, config)

			if len(amazonLinks) > 0 {
				amazonDetails := transformAmazonID(amazonLinks, amazonLinksTitle, amazonImageLinks)

				article := ArticleDetails{
					ArticleURL:    item.Link,
					ArticleTitle:  item.Title,
					Content:       content,
					AmazonDetails: amazonDetails,
				}

				articles = append(articles, article)
			}
		}
	}

	// ここでarticlesを使用して結果を出力
	for _, article := range articles {
		fmt.Println("記事タイトル: ", article.ArticleTitle)
		fmt.Println("記事URL: ", article.ArticleURL)
		fmt.Println("記事内容: ", article.Content)
		fmt.Println(article.AmazonDetails)
		// for _, amazonLink := range article.AmazonDetails {
		// 	fmt.Println(amazonLink.ASIN)
		// 	fmt.Println(amazonLink.URL)
		// 	fmt.Println(amazonLink.URLtitle)
		// 	fmt.Println(amazonLink.ImageURL)
		// }
		fmt.Println("-------------------------")
	}
}
