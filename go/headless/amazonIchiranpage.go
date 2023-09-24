package headless

import (
	"fmt"
	"math/rand"
	"os"
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
	Review             string
	DealURL            string
}

func NewAmazonBrowser() (*AmazonBrowser, error) {
	url := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	page := browser.MustPage()

	// 複数のUser-Agentをスライスに格納
	userAgents := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Safari/605.1.15",
	}

	// ランダムにUser-Agentを選択
	randomUserAgent := userAgents[rand.Intn(len(userAgents))]

	err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: randomUserAgent,
	})
	if err != nil {
		return nil, err
	}

	return &AmazonBrowser{Browser: browser, Page: page}, nil
}

func (ab *AmazonBrowser) GetTopProductUrls() ([]string, error) {
	var topProductUrls []string
	viewIndex := 0

	for len(topProductUrls) < 100 {
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)
		err := ab.Page.Navigate(pageURL)
		if err != nil {
			return nil, fmt.Errorf("error navigating to page: %w", err)
		}

		err = ab.Page.WaitLoad()
		if err != nil {
			return nil, fmt.Errorf("error waiting for page load: %w", err)
		}

		aElements, err := ab.Page.ElementsX(`//div[contains(@class, "gridDisplayGrid")]//a[contains(@class, "a-link-normal") and contains(@class, "DealLink-module")]`)
		if err != nil {
			return nil, fmt.Errorf("error fetching elements: %w", err)
		}

		if len(aElements) == 0 {
			break
		}

		for _, aElement := range aElements {
			url, err := aElement.Attribute("href")
			if err != nil {
				return nil, fmt.Errorf("error getting attribute: %w", err)
			}

			if strings.Contains(*url, "/dp") && !contains(topProductUrls, *url) {
				topProductUrls = append(topProductUrls, *url)
			}
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
		ab.Page.MustNavigate(pageURL)
		ab.Page.MustWaitLoad()

		aElements, err := ab.Page.ElementsX(`//div[contains(@class, "gridDisplayGrid")]//a[contains(@class, "a-link-normal") and contains(@class, "DealLink-module")]`)
		if err != nil {
			return nil, err
		}

		if len(aElements) == 0 { // 新しい要素がない場合、ループを終了
			break
		}

		for _, aElement := range aElements {
			url, err := aElement.Attribute("href")
			if err != nil {
				return nil, err
			}

			if strings.Contains(*url, "/deal") && !contains(topDealUrls, *url) {
				topDealUrls = append(topDealUrls, *url)
			}
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
		ab.Page.MustNavigate(pageURL)
		ab.Page.MustWaitLoad()

		aElements, err := ab.Page.ElementsX(`//div[contains(@class, "gridDisplayGrid")]//a[contains(@class, "a-link-normal") and contains(@class, "DealLink-module")]`)
		if err != nil {
			return nil, nil, err
		}

		if len(aElements) == 0 { // 新しい要素がない場合、ループを終了
			break
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
				if len(topDealUrls) < 50 { // この条件を追加して、topDealUrlsの長さが15未満の場合のみ追加します
					topDealUrls = append(topDealUrls, *url)
				}
			}
		}

		if len(topDealUrls) >= 50 { // 15件のtopDealUrlsに達したら、ループを終了します
			break
		}

		viewIndex += 60
		time.Sleep(1 * time.Second)
	}
	return topProductUrls, topDealUrls, nil
}

func (ab *AmazonBrowser) ProductGetHTMLtags(url string) (ProductPage, error) {
	err := ab.Page.Navigate(url)
	if err != nil {
		return ProductPage{}, fmt.Errorf("error navigating to page: %w", err)
	}

	err = ab.Page.WaitLoad()
	if err != nil {
		return ProductPage{}, fmt.Errorf("error waiting for page load: %w", err)
	}

	// スクリーンショットを取得
	buf := ab.Page.MustScreenshot()
	// スクリーンショットをファイルに保存
	err = os.WriteFile("screenshot.png", buf, 0644)
	if err != nil {
		return ProductPage{}, fmt.Errorf("error saving screenshot: %w", err)
	}

	fmt.Println("商品ページURL: ", url)
	fmt.Println("商品ページのASINを抽出中")

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
	fmt.Println("商品ページのASINを抽出完了: ", asin)

	fmt.Println("商品ページの商品名を抽出中")
	// 商品名取得
	// 商品名の要素がページ上に存在するかどうかを確認するループ
	for i := 0; i < 10; i++ {
		_, err := ab.Page.ElementX(`//span[@id='productTitle']`)
		if err == nil {
			break
		}
		if i == 9 {
			return ProductPage{}, fmt.Errorf("product title not found after multiple attempts")
		}
		time.Sleep(1 * time.Second)
	}

	h1Element, err := ab.Page.ElementX(`//span[@id='productTitle']`)
	if err != nil {
		return ProductPage{}, fmt.Errorf("error fetching product title element: %w", err)
	}
	h1Text, err := h1Element.Text()
	if err != nil {
		return ProductPage{}, fmt.Errorf("error getting text from product title element: %w", err)
	}

	fmt.Println("商品ページの画像URLを抽出中")
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

	fmt.Println("商品ページの割引き情報を抽出中")

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

	fmt.Println("商品ページの値段を抽出中")

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

	fmt.Println("商品ページの商品説明を抽出中")

	// 商品説明テキストか画像を抽出
	extractProductDescription := func() (string, error) {
		Descrip, err := ab.Page.ElementX(`//div[@id='productDescription']`)
		if err != nil {
			return "商品説明なし", err
		}
		DescripText := Descrip.MustText()

		if len(DescripText) > 0 {
			return DescripText, nil
		}
		return "商品説明なし", nil
	}
	descriptionText, _ := extractProductDescription()

	fmt.Println("商品ページのレビューを抽出中")

	// レビューを抽出
	extractProductReview := func() (string, error) {
		ReviewT, err := ab.Page.ElementX(`//div[contains(@class, 'reviewText') or contains(@class, 'review-text-content')]`)
		if err != nil {
			return "レビューなし", err
		}
		ReviewText := ReviewT.MustText()

		if len(ReviewText) > 0 {
			return ReviewText, nil
		}
		return "レビューなし", nil
	}
	reviewText, _ := extractProductReview()

	fmt.Println("商品ページの抽出情報を構造化")

	// structでデータオブジェクト定義
	ab.Product = ProductPage{
		ASIN:               asin,
		ProductName:        h1Text,
		ImageURL:           ImgURL,
		AffURL:             Affurl,
		PriceOff:           priceOfftext,
		Price:              priceText,
		ProductDescription: descriptionText,
		Zaiko:              priceOfftext != "" || priceText != "",
		Review:             reviewText,
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
