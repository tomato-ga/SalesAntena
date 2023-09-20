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
			// まず、URLからコンテンツを抽出する
			doc, rawContent := extractContentFromURL(item.Link, config)

			// 次に、そのドキュメントからAmazonのリンクを抽出
			amazonLinks, amazonImageLinks, amazonLinksTitle := extractAmazonLinksFromDoc(doc, config)
			fmt.Println("extractAmazonLinks--------")
			fmt.Println(amazonLinks)

			// cleanContent関数を使用してコンテンツをクリーンアップ
			cleanedContent := cleanContent(item.Link, rawContent, config.RemoveText, config.RemoveDiv)

			article := ArticleDetails{
				ArticleTitle:  item.Title,
				ArticleURL:    item.Link,
				Content:       cleanedContent,
				AmazonDetails: transformAmazonID(amazonLinks, amazonImageLinks, amazonLinksTitle),
			}

			// articles スライスに追加
			articles = append(articles, article)
		}
	}

	// ここでarticlesを使用して結果を出力
	for _, article := range articles {
		fmt.Println("構造体--------")

		// fmt.Println("記事タイトル: ", article.ArticleTitle)
		// fmt.Println("記事URL: ", article.ArticleURL)
		// fmt.Println("記事内容: ", article.Content)
		fmt.Println(article.AmazonDetails)
		fmt.Println("=========================")
	}
}
