package headless

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type AmazonBrowser struct {
	Browser *rod.Browser
	Page    *rod.Page
	Product ProductPage
}

type ProductPage struct {
	ASIN        string
	ProductName string
	ImageURL    string
	AffURL      string
}

func NewAmazonBrowser() (*AmazonBrowser, error) {
	url := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	page := browser.MustPage()

	customUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"
	err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: customUA,
	})
	if err != nil {
		return nil, err
	}

	return &AmazonBrowser{Browser: browser, Page: page}, nil
}

func (ab *AmazonBrowser) TimesalePage() ([]string, []string, error) {
	var topProductUrls []string
	var topDealUrls []string
	viewIndex := 0

	for len(topProductUrls) < 2000 {
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)
		ab.Page.MustNavigate(pageURL)
		ab.Page.MustWaitLoad()

		aElements, err := ab.Page.ElementsX(`//div[contains(@class, "gridDisplayGrid")]//a[contains(@class, "a-link-normal") and contains(@class, "DealLink-module")]`)
		if err != nil {
			return nil, nil, err
		}

		for _, aElement := range aElements {
			url, err := aElement.Attribute("href")
			if err != nil {
				return nil, nil, err
			}

			// /dpと/dealのURLで分別する
			switch {
			case strings.Contains(*url, "/dp"):
				topProductUrls = append(topProductUrls, *url)
			case strings.Contains(*url, "/deal"):
				topDealUrls = append(topDealUrls, *url)
			}

			if url != nil && *url != "" && !contains(topProductUrls, *url) {
				topProductUrls = append(topProductUrls, *url)
			}
			if len(topProductUrls) >= 30 {
				return topProductUrls, topDealUrls, nil
			}
		}

		viewIndex += 60
		time.Sleep(1 * time.Second)
	}

	return topProductUrls, topDealUrls, nil
}

// TODO 割引率・%を抽出する
func (ab *AmazonBrowser) ProductGetHTMLtags(url string) (ProductPage, error) {
	ab.Page.MustNavigate(url)
	ab.Page.MustWaitLoad()

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

	// 商品名取得
	h1Element, err := ab.Page.Element("h1")
	if err != nil {
		return ProductPage{}, err
	}
	h1Text := h1Element.MustText()

	// 画像URL取得
	extractImageURL := func() (string, error) {
		imgElements, err := ab.Page.Elements("img")
		if err != nil {
			return "", err
		}
		for _, imgElement := range imgElements {
			src, err := imgElement.Attribute("src")
			if err != nil {
				return "", err
			}
			if src != nil && strings.Contains(*src, "media-amazon") {
				return *src, nil
			}
		}
		return "", nil
	}
	ImgURL, err := extractImageURL()
	if err != nil {
		return ProductPage{}, err
	}

	// affurlを生成
	Affurl := url + "&tag=entamenews-22"

	ab.Product = ProductPage{
		ASIN:        asin,
		ProductName: h1Text,
		ImageURL:    ImgURL,
		AffURL:      Affurl,
	}

	return ab.Product, nil
}


// TODO Dealのページ情報は、情報をまとめて一枚のページに出す
func (ab *AmazonBrowser) DealPageURLs(dealURLs []string) ([]string, error) {
	var productURLs []string

	for _, dealURL := range dealURLs {
		ab.Page.MustNavigate(dealURL)
		ab.Page.MustWaitLoad()
		aElements, err := ab.Page.ElementsX(`//div[@id="octopus-dlp-asin-stream"]//li//a[contains(@class, "a-link-normal")]`)
		if err != nil {
			return nil, err
		}

		for _, aElement := range aElements {
			url, err := aElement.Attribute("href")
			if err != nil {
				return nil, err
			}
			if url != nil && *url != "" && !contains(productURLs, *url) {
				completeURL := "https://www.amazon.co.jp" + *url
				productURLs = append(productURLs, completeURL)
			}
		}
	}

	return productURLs, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
