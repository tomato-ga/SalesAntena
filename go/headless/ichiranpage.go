package headless

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Chromeをインストールしてから実行するとOKになる
// wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
// apt install ./google-chrome-stable_current_amd64.deb

func SliceContains(sl []string, item string) bool {
	for _, a := range sl {
		if a == item {
			return true
		}
	}
	return false
}

func SetupBrowser() (*rod.Browser, *rod.Page) {
	url := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	page := browser.MustPage()

	customUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: customUA,
	})

	return browser, page
}

func TimesalePage() []string {
	browser, _ := SetupBrowser()
	topUrl := []string{}

	// 初期のviewIndexの値
	viewIndex := 0

	for len(topUrl) < 1000 {
		// viewIndexを使ってURLを生成
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)
		page := browser.MustPage(pageURL)
		page.MustWaitLoad()

		// XPathを使用して、"Grid-module"を含むクラス名を持つdiv要素を取得
		elements := page.MustElementsX(`//div[contains(@class, "gridDisplayGrid")]`)

		for _, element := range elements {
			hrefs := element.MustElementsX(".//a/@href")
			for _, href := range hrefs {
				url := href.MustText()
				if !SliceContains(topUrl, url) {
					topUrl = append(topUrl, href.MustText())
				}
				if len(topUrl) >= 1000 {
					return topUrl
				}
			}
		}

		// viewIndexの値を更新して次のページに進む
		viewIndex += 60
		// 2秒間スリープ
		time.Sleep(1 * time.Second)
	}
	return topUrl
}

// TODO URLを1000個取得した。