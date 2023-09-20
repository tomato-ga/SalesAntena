package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

type AmazonLinkDetails struct {
	ASIN     string
	URL      string
	URLtitle string
	ImageURL string
}

func extractAmazonURL(rawURL string) (string, error) {
	regexPattern := `vc_url=([^&]+)`
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(rawURL)

	if len(matches) < 2 {
		return "", fmt.Errorf("vc_url parameter not found in %s", rawURL)
	}

	decodedURL, err := url.QueryUnescape(matches[1])
	if err != nil {
		return "", err
	}

	return decodedURL, nil
}

func ensureAmazonAffiliateID(link, affiliateID string) string {
	// リンクがAmazonのものであるか確認
	if strings.Contains(link, "amazon.co.jp") {
		// 既にIDが含まれているか確認
		if !strings.Contains(link, affiliateID) {
			// クエリーパラメータがあるか確認
			if strings.Contains(link, "?") {
				link += "&tag=" + affiliateID
			} else {
				link += "?tag=" + affiliateID
			}
		}
	}
	return link
}

func transformAmazonID(details []AmazonLinkDetails) []AmazonLinkDetails {
	regexForASIN := regexp.MustCompile(`/dp/([A-Z0-9]{10})`)
	regexForTagWithEqual := regexp.MustCompile(`tag=(\w+-22)`)
	regexForTag := regexp.MustCompile(`\w+-22`)

	regexAllDigits := regexp.MustCompile(`^[0-9]{10}$`)
	regexAllLetters := regexp.MustCompile(`^[A-Z]{10}$`)

	var results []AmazonLinkDetails

	for _, detail := range details {
		url := detail.URL
		matches := regexForASIN.FindStringSubmatch(url)
		asin := "ASINなし" // デフォルト値としてASINなしを設定
		if len(matches) > 1 {
			asinCandidate := matches[1]
			if !regexAllDigits.MatchString(asinCandidate) && !regexAllLetters.MatchString(asinCandidate) {
				asin = asinCandidate
			}
		}
		detail.ASIN = asin

		TAG := ""
		matchesTag := regexForTagWithEqual.FindStringSubmatch(url)
		if len(matchesTag) > 1 {
			TAG = matchesTag[1]
		} else {
			matchesTag2 := regexForTag.FindStringSubmatch(url)
			if len(matchesTag2) > 0 {
				TAG = matchesTag2[0]
			}
		}

		if TAG != "" && TAG != "entamenews-22" {
			url = strings.Replace(url, TAG, "entamenews-22", 1)
		}

		newURL := ensureAmazonAffiliateID(url, "entamenews-22")
		detail.URL = newURL

		results = append(results, detail)
	}

	return results
}

func uniqueAmazonURL(amazonURLs []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range amazonURLs {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
