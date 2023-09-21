package main

import (
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type ArticleDetails struct {
	ArticleURL    string
	ArticleTitle  string
	Content       string
	AmazonDetails []AmazonLinkDetails
}

func cleanContent(url string, content string, config FeedConfig) string {
	content = strings.TrimSpace(content)
	re := regexp.MustCompile(`\s+`)
	content = re.ReplaceAllString(content, " ")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err == nil {
		if strings.HasPrefix(url, "https://tokkataro.blog.jp") {
			// Starting from the first removal tag, remove everything thereafter
			for _, removePhrase := range config.RemoveText {
				doc.Find(":contains('" + removePhrase + "')").Each(func(_ int, sel *goquery.Selection) {
					sel.SetText(strings.ReplaceAll(sel.Text(), removePhrase, ""))
					sel.NextAll().Remove()
				})
			}
		} else {
			for _, removePhrase := range config.RemoveText {
				// aタグに含まれているテキストを削除
				doc.Find("a").Each(func(i int, selection *goquery.Selection) {
					if strings.Contains(selection.Text(), removePhrase) {
						selection.Remove()
					}
				})

				// プレーンテキストを削除
				doc.Find(":contains('" + removePhrase + "')").Each(func(i int, s *goquery.Selection) {
					html, _ := s.Html()
					updatedHtml := strings.ReplaceAll(html, removePhrase, "")
					s.SetHtml(updatedHtml)
				})
			}
		}

	}
	content = doc.Text()
	return content
}

func removeUnwantedElements(doc *goquery.Document, config FeedConfig) {
	// テキストの削除
	for _, unwantedText := range config.RemoveText {
		// aタグ内のテキストの処理
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), unwantedText) {
				s.SetText(strings.ReplaceAll(s.Text(), unwantedText, ""))
				s.Remove()
			} else {
				s.SetText(strings.ReplaceAll(s.Text(), unwantedText, ""))
			}
		})

		// その他の要素内のテキストの処理
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), unwantedText) {
				s.SetText(strings.ReplaceAll(s.Text(), unwantedText, ""))
			}
		})
	}

	// 特定のdiv要素の削除
	for _, unwantedDiv := range config.RemoveDiv {
		doc.Find(unwantedDiv).Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
	}
}

func extractContentFromURL(url string, config FeedConfig) (*goquery.Document, string) {
	var content string

	// User Agentの設定
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, content
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error fetching URL:", err)
		return nil, content
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
		return nil, content
	}

	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		content, _ = s.Html()
		content = strings.TrimSpace(content)
		re := regexp.MustCompile(`\s+`)
		content = re.ReplaceAllString(content, " ")
	})

	// 不要な要素の削除
	// removeUnwantedElements(doc, config)
	return doc, content
}

func extractAmazonLinksFromDoc(doc *goquery.Document, config FeedConfig) ArticleDetails {
	var article ArticleDetails

	// Extract amazon links
	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		s.Find("a[href]").Each(func(j int, linkElement *goquery.Selection) {
			link, _ := linkElement.Attr("href")
			text := linkElement.Text()
			isBlacklisted := false
			for _, blackText := range BlackList["Texts"].Black {
				if text == blackText {
					isBlacklisted = true
					break
				}
			}

			amazonDetail := AmazonLinkDetails{}
			if isBlacklisted {
				amazonDetail.URLtitle = "リンクテキストなし"
			} else {
				amazonDetail.URLtitle = text
			}

			switch {
			case strings.Contains(link, "amazon.co.jp"):
				amazonDetail.URL = link
			case strings.Contains(link, "amzn"):
				amazonDetail.URL = link
			case strings.Contains(link, "valuecommerce"):
				amazonURL, err := extractAmazonURL(link)
				if err == nil && (strings.Contains(amazonURL, "amazon.co.jp") || strings.Contains(amazonURL, "amzn")) {
					amazonDetail.URL = amazonURL
				}
			default:
				amazonDetail.URL = "リンクなし"
			}
			article.AmazonDetails = append(article.AmazonDetails, amazonDetail)
		})

		s.Find("img").Each(func(j int, ImageLinkElement *goquery.Selection) {
			imagelink, exists := ImageLinkElement.Attr("src")
			if exists {
				isImageExtension := regexp.MustCompile(`\.(jpg|jpeg|png|gif|bmp|svg|webp)(\?.*)?$`).MatchString(imagelink)
				if isImageExtension && strings.Contains(imagelink, "amazon") {
					if j < len(article.AmazonDetails) {
						article.AmazonDetails[j].ImageURL = imagelink
					} else {
						amazonDetail := AmazonLinkDetails{
							ImageURL: imagelink,
						}
						article.AmazonDetails = append(article.AmazonDetails, amazonDetail)
					}
				}
			}
		})
	})

	return article
}
