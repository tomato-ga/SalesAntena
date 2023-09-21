package main

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

func main() {
	fp := gofeed.NewParser()
	var articles []ArticleDetails
	limitcount := 0

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
			if limitcount >= 3 {
				break
			}

			// URLからコンテンツを抽出する
			doc, rawContent := extractContentFromURL(item.Link, config)

			// ドキュメントからAmazonのリンクを抽出
			art := extractAmazonLinksFromDoc(doc, config)

			amazonDetails := transformAmazonID(art.AmazonDetails)

			// cleanContent関数を使用してコンテンツをクリーンアップ
			cleanedContent := cleanContent(item.Link, rawContent, config)

			// RemoveAllText以降をstring削除する
			for _, rText := range config.RemoveAllText {
				if idx := strings.Index(cleanedContent, rText); idx != -1 {
					cleanedContent = cleanedContent[:idx]
				}
			}

			article := ArticleDetails{
				ArticleTitle:  item.Title,
				ArticleURL:    item.Link,
				Content:       cleanedContent,
				AmazonDetails: amazonDetails,
			}
			articles = append(articles, article)

			// カウント追加
			limitcount++
		}
	}

	for _, artic := range articles {
		fmt.Println("--------------------------------------")
		fmt.Println("タイトル:\n", artic.ArticleTitle)
		fmt.Println("URL:\n", artic.ArticleURL)
		fmt.Println("アマゾンディティール:\n\n", artic.AmazonDetails)
		fmt.Println("記事内容:\n", artic.Content)
		fmt.Println("--------------------------------------")

	}
}
