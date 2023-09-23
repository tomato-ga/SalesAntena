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

	// TODO 単体商品ページと、Dealの商品ページ（複数商品まとめて掲載）は、別の処理をしてDBに保存する
	// 商品ページの情報抽出(Deal以外の商品単体ページの処理)
	for _, url := range topProductUrls {
		if strings.Contains(url, "/dp") {
			Product, err := ab.ProductGetHTMLtags(url)
			if err != nil {
				fmt.Println("ProductGetHTMLタグでエラー発生", err)
			}

			// TODO 商品単体ページの情報を一つずつDBに格納する
			fmt.Println("商品単体ページ： ", Product)
		}
	}

	// TODO Dealの商品まとめでDBに保存する処理
	// Dealページの抽出
	for k, v := range productURLS {
		fmt.Println("Key: ", k)
		for _, value := range v {
			DealProduct, err := ab.ProductGetHTMLtags(value)
			if err != nil {
				fmt.Println("Error in ProductGetHTMLtags for DealProduct:", err)
			}
			fmt.Println("/dealのページの商品一覧: ", DealProduct)
		}
	}
}
