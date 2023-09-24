package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ProductPage struct {
	ASIN               string
	ProductName        string
	ImageURL           string
	AffURL             string
	PriceOff           string
	Price              string
	ProductDescription string
	Zaiko              bool
	Review             string
	DealURL            string
}

type AmazonBrowser struct {
	Client    *http.Client
	UserAgent string
}

func NewAmazonBrowser() *AmazonBrowser {
	// ランダムにUser-Agentを選択
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36",
		// 他のUser-Agentも追加可能
	}
	randomUserAgent := userAgents[rand.Intn(len(userAgents))]

	client := &http.Client{}
	return &AmazonBrowser{Client: client, UserAgent: randomUserAgent}
}

func (ab *AmazonBrowser) GetTopProductUrls() ([]string, error) {
	var topProductUrls []string
	viewIndex := 0

	for len(topProductUrls) < 100 {
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)

		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", ab.UserAgent)

		resp, err := ab.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}

		doc.Find("div.gridDisplayGrid a.a-link-normal.DealLink-module").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/dp") && !contains(topProductUrls, href) {
				topProductUrls = append(topProductUrls, "https://www.amazon.co.jp"+href)
			}
		})

		if len(topProductUrls) == 0 {
			break
		}

		viewIndex += 60
		time.Sleep(1 * time.Second)
	}
	return topProductUrls, nil
}

func (ab *AmazonBrowser) GetTopDealUrls() ([]string, error) {
	var topDealUrls []string
	viewIndex := 0

	for len(topDealUrls) < 40 {
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)

		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", ab.UserAgent)

		resp, err := ab.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}

		doc.Find("div.gridDisplayGrid a.a-link-normal.DealLink-module").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && strings.Contains(href, "/deal") && !contains(topDealUrls, href) {
				topDealUrls = append(topDealUrls, "https://www.amazon.co.jp"+href)
			}
		})

		if len(topDealUrls) == 0 {
			break
		}

		viewIndex += 60
		time.Sleep(1 * time.Second)
	}
	return topDealUrls, nil
}

func (ab *AmazonBrowser) TimesalePage() ([]string, []string, error) {
	var topProductUrls []string
	var topDealUrls []string
	viewIndex := 0

	for {
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)

		req, err := http.NewRequest("GET", pageURL, nil)
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("User-Agent", ab.UserAgent)

		resp, err := ab.Client.Do(req)
		if err != nil {
			return nil, nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, nil, err
		}

		doc.Find("div.gridDisplayGrid a.a-link-normal.DealLink-module").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				switch {
				case strings.Contains(href, "/dp") && !contains(topProductUrls, href):
					topProductUrls = append(topProductUrls, "https://www.amazon.co.jp"+href)
				case strings.Contains(href, "/deal") && !contains(topDealUrls, href) && len(topDealUrls) < 50:
					topDealUrls = append(topDealUrls, "https://www.amazon.co.jp"+href)
				}
			}
		})

		if len(topDealUrls) >= 50 || len(topProductUrls) == 0 {
			break
		}

		viewIndex += 60
		time.Sleep(1 * time.Second)
	}
	return topProductUrls, topDealUrls, nil
}

func (ab *AmazonBrowser) ProductGetHTMLtags(url string) (ProductPage, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ProductPage{}, err
	}
	req.Header.Set("User-Agent", ab.UserAgent)

	resp, err := ab.Client.Do(req)
	if err != nil {
		return ProductPage{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ProductPage{}, err
	}

	// ASIN抽出
	extractASIN := func(u string) string {
		regexForASIN := regexp.MustCompile(`/dp/([A-Z0-9]{10})`)
		matches := regexForASIN.FindStringSubmatch(u)
		if len(matches) > 1 {
			return matches[1]
		}
		return ""
	}
	asin := extractASIN(url)

	// 商品名の取得
	productName := doc.Find("span#productTitle").Text()

	// 画像URLの取得
	imgURL, _ := doc.Find("div.imgTagWrapper img").Attr("src")

	// 割引きを抽出
	priceOff := doc.Find("span.a-offscreen").Text()

	// 値段を抽出
	price := doc.Find("span.a-offscreen").Text()

	// 商品説明テキストか画像を抽出
	description := doc.Find("div#productDescription").Text()

	// レビューを抽出
	review := doc.Find("div.reviewText, div.review-text-content").Text()

	// affurlを生成
	affURL := url + "&tag=entamenews-22"

	// structでデータオブジェクト定義
	product := ProductPage{
		ASIN:               asin,
		ProductName:        productName,
		ImageURL:           imgURL,
		AffURL:             affURL,
		PriceOff:           priceOff,
		Price:              price,
		ProductDescription: description,
		Zaiko:              priceOff != "" || price != "",
		Review:             review,
	}

	return product, nil
}

func (ab *AmazonBrowser) DealPageURLs(dealURLs []string) (map[string][]string, error) {
	dealToProducts := make(map[string][]string)

	for _, dealURL := range dealURLs {
		req, err := http.NewRequest("GET", dealURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", ab.UserAgent)

		resp, err := ab.Client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}

		var productURLsForDeal []string
		doc.Find("div#octopus-dlp-asin-stream li a.a-link-normal").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists && href != "" && !contains(productURLsForDeal, href) {
				completeURL := "https://www.amazon.co.jp" + href
				productURLsForDeal = append(productURLsForDeal, completeURL)
			}
		})

		dealToProducts[dealURL] = productURLsForDeal
	}
	return dealToProducts, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
