package main

import (
	"Salesscrape/headless"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func main() {
	fmt.Println("---スタート---")

	// 0.7秒から1.5秒の間のランダムなスリープ時間を取得
	randomFloat := 0.7 + rand.Float64()*(1.5-0.7)
	sleepDuration := time.Duration(randomFloat * float64(time.Second))
	// スリープ
	time.Sleep(sleepDuration)

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
			// スリープ
			time.Sleep(sleepDuration)
			// TODO 商品単体ページの情報を一つずつDBに格納する
			fmt.Println("商品単体ページ： ", Product)
		} else {
			continue
		}
	}

	// TODO Dealの商品まとめでDBに保存する処理
	// Dealページの抽出
	for k, v := range productURLS {
		fmt.Println("Key: ", k)
		for _, value := range v {
			DealProduct, _ := ab.ProductGetHTMLtags(value)
			time.Sleep(sleepDuration)

			fmt.Println("/dealのページの商品一覧: ", DealProduct)
		}
	}
}
