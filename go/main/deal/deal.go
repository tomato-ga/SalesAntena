package main

import (
	"Salesscrape/headless"
	"fmt"
)

func main() {
	fmt.Println("---Dealページの処理を開始---")

	ab, err := headless.NewAmazonBrowser()
	if err != nil {
		fmt.Println("Error create AmazonBrowser", err)
		return
	}

	topDealUrls, err := ab.GetTopDealUrls()
	if err != nil {
		fmt.Println("Error fetching timesale page", err)
		return
	}

	fmt.Println("---DEAL---")
	fmt.Println(len(topDealUrls))

	productURLS, err := ab.DealPageURLs(topDealUrls)
	if err != nil {
		fmt.Println("Error fetching timesale page", err)
		return
	}

	ProcessDealPages(ab, productURLS)
}

func ProcessDealPages(ab *headless.AmazonBrowser, productURLS map[string][]string) {
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
