package main

import (
	"Salesscrape/headless"
	"fmt"
)

func main() {
	topUrl, _ := headless.TimesalePage()

	fmt.Println(topUrl)
}
