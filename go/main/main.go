package main

import (
	"Salesscrape/headless"
	"fmt"
	"strings"
)

func main() {
	fmt.Println("---スタート---")

	ab, err := headless.NewAmazonBrowser()
	if err != nil {
		fmt.Println("Error create AmazonBrowser", err)
		return
	}

	topProductUrls, topDealUrls, err := ab.TimesalePage()
	if err != nil {
		fmt.Println("Error fetching timesale page", err)
		return
	}

	fmt.Println("---Product---")
	fmt.Println(len(topProductUrls))
	fmt.Println("---DEAL---")
	fmt.Println(len(topDealUrls))

	// dealページから商品ページURL/dpを取得 -> mapでdealURL[商品URL]が戻り値になる
	productURLS, err := ab.DealPageURLs(topDealUrls)
	if err != nil {
		fmt.Println("Error fetching timesale page", err)
		return
	}

	// 商品ページの情報抽出(Deal以外の商品単体ページの処理)
	for _, url := range topProductUrls {
		if strings.Contains(url, "/dp") {
			Product, err := ab.ProductGetHTMLtags(url)
			if err != nil {
				fmt.Println("ProductGetHTMLタグでエラー発生", err)
			}

			fmt.Println("商品単体ページ： ", Product.ProductName)

			// DynamoDBに保存する
			err = headless.PutItemtoDynamoDB(Product)
			if err != nil {
				fmt.Println("DynamoDBの保存でエラーが出ました", err)
			}

		}
	}

	// Dealページの抽出
	for k, v := range productURLS {
		fmt.Println("Key: ", k)
		for _, value := range v {
			DealProduct, err := ab.ProductGetHTMLtags(value)
			DealProduct.DealURL = k

			if err != nil {
				fmt.Println("Error in ProductGetHTMLtags for DealProduct:", err)
			}
			fmt.Println("/dealのページの商品一覧: ", DealProduct)

			// DynamoDBに保存する
			err = headless.PutItemtoDynamoDB(DealProduct)
			if err != nil {
				fmt.Println("Deal商品のDynamoDBの保存でエラーがでました", err)
			}
		}
	}
}

// TODO 商品単体ページとDealページは別ファイルで実行するのがよさそう
