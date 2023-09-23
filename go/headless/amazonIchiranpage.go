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
	ASIN               string
	ProductName        string
	ImageURL           string
	AffURL             string
	PriceOff           string
	Price              string
	ProductDescription string
	Zaiko              bool
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
			case strings.Contains(*url, "/dp") && !contains(topProductUrls, *url):
				topProductUrls = append(topProductUrls, *url)
			case strings.Contains(*url, "/deal") && !contains(topDealUrls, *url):
				topDealUrls = append(topDealUrls, *url)
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

	extractImageURL := func() (string, error) {
		// divタグのクラスがimgTagWrapperで、その子要素として<img>タグが存在する要素を検索
		imgElements, err := ab.Page.ElementsX(`//div[@class="imgTagWrapper"]/img`)
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

	// 割引きを抽出
	extractPriceoff := func() (string, error) {
		// 割引きの正規表現
		rePrice := regexp.MustCompile(`-\d+%`)

		priceOffElements, err := ab.Page.ElementsX(`//div[contains(@class, 'a-section')]/span`)
		if err != nil {
			return "", err
		}

		for _, priceOffElement := range priceOffElements {
			priceOffText := priceOffElement.MustText()
			if rePrice.MatchString(priceOffText) {
				return priceOffText, nil
			}
		}
		return "割引きなし", nil
	}
	priceOfftext, _ := extractPriceoff()

	// 値段を抽出
	extractPrice := func() (string, error) {
		rePrice := regexp.MustCompile(`￥([\d,]+)`)

		priceElements, err := ab.Page.ElementsX(`//span[@class='a-offscreen']`)
		if err != nil {
			return "", err
		}

		for _, priceElement := range priceElements {
			priceText := priceElement.MustText()
			if rePrice.MatchString(priceText) {
				return priceText, nil
			}
		}
		return "値段なし", nil
	}
	priceText, _ := extractPrice()

	// 商品説明テキストか画像を抽出
	extractProductDescription := func() (string, error) {
		Descrip, err := ab.Page.ElementX(`//div[@id='productDescription']`)
		if err != nil {
			return "", err
		}
		DescripText := Descrip.MustText()

		if len(DescripText) > 0 {
			return DescripText, nil
		}
		return "", nil
	}
	descriptionText, _ := extractProductDescription()

	ab.Product = ProductPage{
		ASIN:               asin,
		ProductName:        h1Text,
		ImageURL:           ImgURL,
		AffURL:             Affurl,
		PriceOff:           priceOfftext,
		Price:              priceText,
		ProductDescription: descriptionText,
		Zaiko:              priceOfftext != "" || priceText != "",
	}

	return ab.Product, nil
}

func (ab *AmazonBrowser) DealPageURLs(dealURLs []string) (map[string][]string, error) {
	dealToProducts := make(map[string][]string)

	for _, dealURL := range dealURLs {
		ab.Page.MustNavigate(dealURL)
		ab.Page.MustWaitLoad()
		aElements, err := ab.Page.ElementsX(`//div[@id="octopus-dlp-asin-stream"]//li//a[contains(@class, "a-link-normal")]`)
		if err != nil {
			return nil, err
		}

		var productURLsForDeal []string
		for _, aElement := range aElements {
			url, err := aElement.Attribute("href")
			if err != nil {
				return nil, err
			}
			if url != nil && *url != "" && !contains(productURLsForDeal, *url) {
				completeURL := "https://www.amazon.co.jp" + *url
				productURLsForDeal = append(productURLsForDeal, completeURL)
			}
		}
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
