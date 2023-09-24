package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("---スタート---")

	ab := NewAmazonBrowser()

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
			if err != nil {
				fmt.Println("Deal商品のDynamoDBの保存でエラーがでました", err)
			}
		}
	}
}
