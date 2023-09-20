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

func cleanContent(url string, content string, removeText []string, removeDiv []string) string {
	content = strings.TrimSpace(content)
	re := regexp.MustCompile(`\s+`)
	content = re.ReplaceAllString(content, " ")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err == nil {
		if strings.HasPrefix(url, "https://tokkataro.blog.jp") {
			// Starting from the first removal tag, remove everything thereafter
			for _, removePhrase := range removeText {
				doc.Find(":contains('" + removePhrase + "')").Each(func(_ int, sel *goquery.Selection) {
					sel.SetText(strings.ReplaceAll(sel.Text(), removePhrase, ""))
					sel.NextAll().Remove()
				})
			}
		} else {
			for _, removePhrase := range removeText {
				doc.Find("a").Each(func(i int, selection *goquery.Selection) {
					if strings.Contains(selection.Text(), removePhrase) {
						selection.SetText(strings.ReplaceAll(selection.Text(), removePhrase, ""))
						// selection.Remove()
					} else {
						selection.SetText(strings.ReplaceAll(selection.Text(), removePhrase, ""))
					}
				})
			}

			// removeDiv's processing is applied to all URLs
			for _, selector := range removeDiv {
				doc.Find(selector).Each(func(i int, selection *goquery.Selection) {
					selection.Remove()
				})
			}
		}
	}

	content = doc.Text()
	return content
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

	return doc, content
}

// TODO データの流れが分かっていないから処理がこんがらがる
func extractAmazonLinksFromDoc(doc *goquery.Document, config FeedConfig) ([]string, []string, []string) {
	var amazonLinks []string
	var amazonLinksTitle []string
	var amazonImageLinks []string

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
			default:
				amazonLinks = append(amazonLinks, "リンクなし")
			}

			// Ensure all slices have the same length
			for len(amazonLinks) > len(amazonImageLinks) {
				amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
			}
			for len(amazonLinks) > len(amazonLinksTitle) {
				amazonLinksTitle = append(amazonLinksTitle, "リンクテキストなし")
			}
		})

		s.Find("img").Each(func(j int, ImageLinkElement *goquery.Selection) {
			imagelink, exists := ImageLinkElement.Attr("src")
			if exists {
				isImageExtension := regexp.MustCompile(`\.(jpg|jpeg|png|gif|bmp|svg|webp)(\?.*)?$`).MatchString(imagelink)
				if isImageExtension && strings.Contains(imagelink, "amazon") {
					amazonImageLinks = append(amazonImageLinks, imagelink)
				} else {
					amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
				}
			} else {
				amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
			}

			// Ensure all slices have the same length
			for len(amazonImageLinks) > len(amazonLinks) {
				amazonLinks = append(amazonLinks, "リンクなし")
			}
			for len(amazonImageLinks) > len(amazonLinksTitle) {
				amazonLinksTitle = append(amazonLinksTitle, "リンクテキストなし")
			}
		})
	})

	return amazonLinks, amazonImageLinks, amazonLinksTitle
}

// func extractAmazonLinks(url string, config FeedConfig) (string, []string, []string, []string) {
// 	var content string
// 	var amazonLinks []string
// 	var amazonLinksTitle []string
// 	var amazonImageLinks []string

// 	// User Agentの設定
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		log.Println("Error creating request:", err)
// 		return content, amazonLinks, amazonImageLinks, amazonLinksTitle
// 	}
// 	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Println("Error fetching URL:", err)
// 		return content, amazonLinks, amazonImageLinks, amazonLinksTitle
// 	}
// 	defer resp.Body.Close()

// 	var reader io.Reader
// 	if resp.Header.Get("Content-Type") == "text/html; charset=euc-jp" {
// 		reader = transform.NewReader(resp.Body, japanese.EUCJP.NewDecoder())
// 	} else {
// 		reader = resp.Body
// 	}

// 	doc, err := goquery.NewDocumentFromReader(reader)
// 	if err != nil {
// 		log.Println("Error reading the document:", err)
// 		return content, amazonLinks, amazonImageLinks, amazonLinksTitle
// 	}

// 	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
// 		content, _ = s.Html()
// 		content = strings.TrimSpace(content)
// 		re := regexp.MustCompile(`\s+`)
// 		content = re.ReplaceAllString(content, " ")
// 	})

// 	// Extract amazon links
// 	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
// 		s.Find("a[href]").Each(func(j int, linkElement *goquery.Selection) {
// 			link, _ := linkElement.Attr("href")

// 			text := linkElement.Text()
// 			isBlacklested := false
// 			for _, blackText := range BlackList["Texts"].Black {
// 				if text == blackText {
// 					isBlacklested = true
// 					break
// 				}
// 			}

// 			if isBlacklested {
// 				amazonLinksTitle = append(amazonLinksTitle, "リンクテキストなし")
// 			} else {
// 				amazonLinksTitle = append(amazonLinksTitle, text)
// 			}

// 			switch {
// 			case strings.Contains(link, "amazon.co.jp"):
// 				amazonLinks = append(amazonLinks, link)
// 			case strings.Contains(link, "amzn"):
// 				amazonLinks = append(amazonLinks, link)
// 			case strings.Contains(link, "valuecommerce"):
// 				amazonURL, err := extractAmazonURL(link)
// 				if err == nil && (strings.Contains(amazonURL, "amazon.co.jp") || strings.Contains(amazonURL, "amzn")) {
// 					amazonLinks = append(amazonLinks, amazonURL)
// 				}
// 			}
// 		})

// 		s.Find("img").Each(func(j int, ImageLinkElement *goquery.Selection) {
// 			imagelink, exists := ImageLinkElement.Attr("src")
// 			if exists {
// 				// 拡張子の確認
// 				isImageExtension := regexp.MustCompile(`\.(jpg|jpeg|png|gif|bmp|svg|webp)(\?.*)?$`).MatchString(imagelink)
// 				if isImageExtension && strings.Contains(imagelink, "amazon") {
// 					amazonImageLinks = append(amazonImageLinks, imagelink)
// 				} else {
// 					amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
// 				}
// 			}
// 		})
// 	})

// 	maxLength := max(len(amazonLinks), len(amazonImageLinks), len(amazonLinksTitle))

// 	for len(amazonLinks) < maxLength {
// 		amazonLinks = append(amazonLinks, "リンクなし")
// 	}
// 	for len(amazonImageLinks) < maxLength {
// 		amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
// 	}
// 	for len(amazonLinksTitle) < maxLength {
// 		amazonLinksTitle = append(amazonLinksTitle, "リンクテキストなし")
// 	}

// 	return content, amazonLinks, amazonImageLinks, amazonLinksTitle
// }

// func max(nums ...int) int {
// 	max := nums[0]
// 	for _, n := range nums {
// 		if n > max {
// 			max = n
// 		}
// 	}
// 	return max
// }
