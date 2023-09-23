package main

import (
	"Salesscrape/headless"
	"fmt"
	"math/rand"
	"time"
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

	// dealページから商品ページURLを取得
	productURLS, err := ab.DealPageURLs(topDealUrls)
	if err != nil {
		fmt.Println("Error fetching timesale page", err)
		return
	}

	// 商品ページURLを合体させる→合体必要ない
	// TODO 単体商品ページと、Dealの商品ページ（複数商品まとめて掲載）は、別の処理をしてDBに保存する
	mixedProductURLs := append(topProductUrls, productURLS...)

	// 商品ページの情報抽出
	for _, url := range mixedProductURLs {
		Product, err := ab.ProductGetHTMLtags(url)

		// 0.7秒から1.5秒の間のランダムなスリープ時間を取得
		randomFloat := 0.7 + rand.Float64()*(1.5-0.7)
		sleepDuration := time.Duration(randomFloat * float64(time.Second))

		// スリープ
		time.Sleep(sleepDuration)

		if err != nil {
			fmt.Println("Error fetching H1 text", err)
			continue
		}

		fmt.Println(Product)
	}

}
