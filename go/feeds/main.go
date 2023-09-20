package main

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

func main() {
	fp := gofeed.NewParser()
	var articles []ArticleDetails

	for feedURL, config := range RssListsMap {
		feed, _ := fp.ParseURL(feedURL)

		for _, item := range feed.Items {
			// URLからコンテンツを抽出する
			doc, rawContent := extractContentFromURL(item.Link, config)
			
			// ドキュメントからAmazonのリンクを抽出
			art := extractAmazonLinksFromDoc(doc, config)

			amazonDetails := transformAmazonID(art.AmazonDetails)
			// cleanContent関数を使用してコンテンツをクリーンアップ
			cleanedContent := cleanContent(item.Link, rawContent, config.RemoveText, config.RemoveDiv)

			article := ArticleDetails{
				ArticleTitle:  item.Title,
				ArticleURL:    item.Link,
				Content:       cleanedContent,
				AmazonDetails: amazonDetails,
			}
			articles = append(articles, article)
		}
	}
	for _, artic := range articles {
		fmt.Println("--------------------------------------")
		fmt.Println("タイトル", artic.ArticleTitle)
		fmt.Println("URL", artic.ArticleURL)
		fmt.Println("アマゾンディティール", artic.AmazonDetails)
		fmt.Println("記事内容", artic.Content)
		fmt.Println("--------------------------------------")

	}
}
