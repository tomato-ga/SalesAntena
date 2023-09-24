package main

import (
	"Salesscrape/headless"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("---商品単体ページの処理を開始---")

	ab, err := headless.NewAmazonBrowser()
	if err != nil {
		fmt.Println("Error creating AmazonBrowser:", err)
		return
	}

	topProductUrls, err := ab.GetTopProductUrls()
	if err != nil {
		fmt.Println("Error fetching top product URLs:", err)
		return
	}

	fmt.Println("---Product---")
	fmt.Println("商品単体ページは", len(topProductUrls), "件あります")

	ProcessProductPages(ab, topProductUrls)
}

func ProcessProductPages(ab *headless.AmazonBrowser, topProductUrls []string) {

	fmt.Println("商品ページを抽出スタートします")

	for _, url := range topProductUrls {
		if strings.Contains(url, "/dp") {
			Product, err := ab.ProductGetHTMLtags(url)
			fmt.Println("商品ページを抽出完了しました")

			if err != nil {
				fmt.Println("Error in ProductGetHTMLtags:", err)
				continue
			}

			fmt.Println("商品単体ページの情報→： ", Product.ProductName)

			// DynamoDBに保存する
			err = headless.PutItemtoDynamoDB(Product)
			if err != nil {
				fmt.Println("Error saving to DynamoDB:", err)
			}
		}
	}
}
