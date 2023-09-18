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

// TODO RSS URLからソースを取得し、ASIN amazonURL amazonTitleを取得してから、いらないものを削除する方向でやってみる
// TODO AmazonFetchもなくす

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
	regexPattern := `vc_url=([^&]+)`
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(rawURL)

	if len(matches) < 2 {
		return "", fmt.Errorf("vc_url parameter not found in %s", rawURL)
	}

	decodedURL, err := url.QueryUnescape(matches[1])
	if err != nil {
		return "", err
	}

	return decodedURL, nil
}

func ensureAmazonAffiliateID(link, affiliateID string) string {
	// リンクがAmazonのものであるか確認
	if strings.Contains(link, "amazon.co.jp") {
		// 既にIDが含まれているか確認
		if !strings.Contains(link, affiliateID) {
			// クエリーパラメータがあるか確認
			if strings.Contains(link, "?") {
				link += "&tag=" + affiliateID
			} else {
				link += "?tag=" + affiliateID
			}
		}
	}
	return link
}

func transformAmazonID(urls []string, urlsTitle []string, imageLinks []string) []AmazonLinkDetails {
	// if len(urls) == 0 || len(urlsTitle) == 0 || len(imageLinks) == 0 {
	// 	log.Printf("Empty slice detected: urls(%d), urlsTitle(%d), imageLinks(%d)", len(urls), len(urlsTitle), len(imageLinks))
	// 	return nil
	// }
	regexForASIN := regexp.MustCompile(`/([A-Z0-9]{10})`)
	regexForTagWithEqual := regexp.MustCompile(`tag=(\w+-22)`)
	regexForTag := regexp.MustCompile(`\w+-22`)

	regexAllDigits := regexp.MustCompile(`^[0-9]{10}$`)
	regexAllLetters := regexp.MustCompile(`^[A-Z]{10}$`)

	var results []AmazonLinkDetails

	for index, url := range urls {
		matches := regexForASIN.FindStringSubmatch(url)
		if len(matches) > 1 {
			asin := matches[1]

			if regexAllDigits.MatchString(asin) || regexAllLetters.MatchString(asin) {
				continue
			}

			TAG := ""

			matchesTag := regexForTagWithEqual.FindStringSubmatch(url)
			if len(matchesTag) > 1 {
				TAG = matchesTag[1]
			} else {
				matchesTag2 := regexForTag.FindStringSubmatch(url)
				if len(matchesTag2) > 0 {
					TAG = matchesTag2[0]
				}
			}

			if TAG != "" && TAG != "entamenews-22" {
				url = strings.Replace(url, TAG, "entamenews-22", 1)
			}

			newURL := ensureAmazonAffiliateID(url, "entamenews-22")

			imageURL := imageLinks[index]
			title := urlsTitle[index]

			results = append(results, AmazonLinkDetails{
				ASIN:     asin,
				URL:      newURL,
				URLtitle: title,
				ImageURL: imageURL,
			})
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
					sel.Remove()
					sel.NextAll().Remove()
				})
			}

		} else {
			for _, removePhrase := range removeText {
				doc.Find("a").Each(func(i int, selection *goquery.Selection) {
					if strings.Contains(selection.Text(), removePhrase) {
						selection.Remove()
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
		content, _ = doc.Find("body").First().Html()
		re := regexp.MustCompile(`<[^>]*>`)
		content = re.ReplaceAllString(content, "") // Remove all HTML tags
		content = strings.TrimSpace(content)
		content = re.ReplaceAllString(content, " ")
	}

	return content
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

	doc.Find(config.Selector).Each(func(i int, s *goquery.Selection) {
		// RemoveDivで指定された要素を削除
		for _, divToRemove := range config.RemoveDiv {
			s.Find(divToRemove).Remove()
		}

		content = s.Text()

		// RemoveTextの各項目に基づいて、該当テキストを単に削除
		for _, removePhrase := range config.RemoveText {
			content = strings.ReplaceAll(content, removePhrase, "")
		}

		content = strings.TrimSpace(content)
		re := regexp.MustCompile(`\s+`)
		content = re.ReplaceAllString(content, " ")
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
			if exists {
				// 拡張子の確認
				isImageExtension := regexp.MustCompile(`\.(jpg|jpeg|png|gif|bmp|svg|webp)(\?.*)?$`).MatchString(imagelink)
				if isImageExtension && strings.Contains(imagelink, "amazon") {
					amazonImageLinks = append(amazonImageLinks, imagelink)
				} else {
					amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
				}
			}
		})
	})

	maxLength := max(len(amazonLinks), len(amazonImageLinks), len(amazonLinksTitle))

	for len(amazonLinks) < maxLength {
		amazonLinks = append(amazonLinks, "リンクなし")
	}
	for len(amazonImageLinks) < maxLength {
		amazonImageLinks = append(amazonImageLinks, "アマゾン画像なし")
	}
	for len(amazonLinksTitle) < maxLength {
		amazonLinksTitle = append(amazonLinksTitle, "リンクテキストなし")
	}

	return content, amazonLinks, amazonImageLinks, amazonLinksTitle
}

func max(nums ...int) int {
	max := nums[0]
	for _, n := range nums {
		if n > max {
			max = n
		}
	}
	return max
}

func main() {
	fp := gofeed.NewParser()
	var articles []ArticleDetails

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
			rawContent, amazonLinks, amazonImageLinks, amazonLinksTitle := extractAmazonLinks(item.Link, config)

			// cleanContent関数を使用してコンテンツをクリーンアップ
			cleanedContent := cleanContent(item.Link, rawContent, config.RemoveText, config.RemoveDiv)

			article := ArticleDetails{
				ArticleTitle:  item.Title,
				ArticleURL:    item.Link,
				Content:       cleanedContent,
				AmazonDetails: transformAmazonID(amazonLinks, amazonImageLinks, amazonLinksTitle),
			}

			// AmazonDetailsをフィルタリング
			var filteredAmazonDetails []AmazonLinkDetails
			for _, detail := range article.AmazonDetails {
				shouldRemove := false
				for _, removePhrase := range config.RemoveText {
					if strings.Contains(detail.ASIN, removePhrase) ||
						strings.Contains(detail.URL, removePhrase) ||
						strings.Contains(detail.URLtitle, removePhrase) {
						shouldRemove = true
						break
					}
				}
				if !shouldRemove {
					filteredAmazonDetails = append(filteredAmazonDetails, detail)
				}
			}
			article.AmazonDetails = filteredAmazonDetails

			// articles スライスに追加
			articles = append(articles, article)
		}
	}

	// ここでarticlesを使用して結果を出力
	for _, article := range articles {
		fmt.Println("記事タイトル: ", article.ArticleTitle)
		fmt.Println("記事URL: ", article.ArticleURL)
		fmt.Println("記事内容: ", article.Content)
		fmt.Println(article.AmazonDetails)

		fmt.Println("=========================")
	}
}
