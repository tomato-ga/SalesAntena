package main

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

func main() {
	fp := gofeed.NewParser()
	var articles []ArticleDetails

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
//TODO : amazon urlを取得する前に、div要素を取得するだけの関数を用意する

			rawContent, amazonLinks, amazonImageLinks, amazonLinksTitle := extractAmazonLinks(item.Link, config)

			// cleanContent関数を使用してコンテンツをクリーンアップ
			cleanedContent := cleanContent(item.Link, rawContent, config.RemoveText, config.RemoveDiv)

			article := ArticleDetails{
				ArticleTitle:  item.Title,
				ArticleURL:    item.Link,
				Content:       cleanedContent,
				AmazonDetails: transformAmazonID(amazonLinks, amazonImageLinks, amazonLinksTitle),
			}

			// AmazonDetailsをフィルタリング ここいるか？？？
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
