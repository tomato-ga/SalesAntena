package headless

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// Chromeをインストールしてから実行するとOKになる
// wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
// apt install ./google-chrome-stable_current_amd64.deb


func Heads() {
	url := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()

	page := browser.MustPage("https://www.amazon.co.jp/gp/goldbox")
	page.MustWaitLoad()

	// XPathを使用して、"Grid-module"を含むクラス名を持つdiv要素を取得
	elements := page.MustElementsX(`//div[contains(@class, "gridDisplayGrid")]`)

	for _, element := range elements {
		fmt.Println(element.MustText())

		hrefs := element.MustElementsX(".//a/@href")
		for _, href := range hrefs {
			fmt.Println(href.MustText())
		}
	}
}

// TODO ページネーションの全てのページで、商品URLをスライスで取得する→個々のURLにアクセスして、価格情報等を取得する
