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

func SetupBrowser() (*rod.Browser, *rod.Page, error) {
	url := launcher.New().MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	page := browser.MustPage()

	customUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36"
	err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: customUA,
	})
	if err != nil {
		return nil, nil, err
	}

	return browser, page, nil
}

func TimesalePage() ([]string, error) {
	browser, _, err := SetupBrowser()
	if err != nil {
		return nil, err
	}

	topUrl := make([]string, 0, 2000)
	viewIndex := 0

	for len(topUrl) < 2000 {
		pageURL := fmt.Sprintf("https://www.amazon.co.jp/gp/goldbox?viewIndex=%d", viewIndex)
		page := browser.MustPage(pageURL)
		page.MustWaitLoad()

		elements, err := page.ElementsX(`//div[contains(@class, "gridDisplayGrid")]`)
		if err != nil {
			return nil, err
		}

		for _, element := range elements {
			hrefs, err := element.ElementsX(".//a/@href")
			if err != nil {
				return nil, err
			}

			for _, href := range hrefs {
				url := href.MustText()
				if !contains(topUrl, url) {
					topUrl = append(topUrl, url)
				}
				if len(topUrl) >= 2000 {
					return topUrl, nil
				}
			}
		}

		viewIndex += 60
		time.Sleep(1 * time.Second)
	}
	return topUrl, nil
}

func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}