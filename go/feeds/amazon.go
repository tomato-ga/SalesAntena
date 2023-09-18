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

func transformAmazonID(urls []string, urlsTitle []string, imageLinks []string) []AmazonLinkDetails {
	// if len(urls) == 0 || len(urlsTitle) == 0 || len(imageLinks) == 0 {
	// 	log.Printf("Empty slice detected: urls(%d), urlsTitle(%d), imageLinks(%d)", len(urls), len(urlsTitle), len(imageLinks))
	// 	return nil
	// }
	regexForASIN := regexp.MustCompile(`/([A-Z0-9]{10})`)
	regexForTagWithEqual := regexp.MustCompile(`tag=(\w+-22)`)
	regexForTag := regexp.MustCompile(`\w+-22`)

	regexAllDigits := regexp.MustCompile(`^[0-9]{10}$`)
	regexAllLetters := regexp.MustCompile(`^[A-Z]{10}$`)

	var results []AmazonLinkDetails

	for index, url := range urls {
		matches := regexForASIN.FindStringSubmatch(url)
		if len(matches) > 1 {
			asin := matches[1]

			if regexAllDigits.MatchString(asin) || regexAllLetters.MatchString(asin) {
				continue
			}

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

			imageURL := imageLinks[index]
			title := urlsTitle[index]

			results = append(results, AmazonLinkDetails{
				ASIN:     asin,
				URL:      newURL,
				URLtitle: title,
				ImageURL: imageURL,
			})
		}
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