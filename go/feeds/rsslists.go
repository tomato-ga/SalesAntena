package main

var RssLists = []string{
	"http://tokkaban.com/feed",
	// "https://tokkataro.blog.jp/atom.xml",
	// "https://gekiyasu-gekiyasu.doorblog.jp/index.rdf",
	// "http://buy.livedoor.biz/index.rdf",
	"http://ks4402.blog94.fc2.com/?xml",
	"http://tokkagekiyasu.blog105.fc2.com/?xml",
	"http://digitalcamera3.blog56.fc2.com/?xml",
	"http://tvxtv.blog120.fc2.com/?xml",
	"https://webtokubai.blog.fc2.com/?xml",
	"http://feeds.feedburner.com/919cc",
	"https://iitokimowaruitokimo.com/feed",
	"https://gekiyasu.blog/feed",
	"https://blog.919.bz/index.rdf",
	"https://fanblogs.jp/nightfly/index1_0.rdf",
	"https://www.nagano-life.net/blog/feed",
	"https://gekiyasutoka.com/index.rdf",
	"https://yasu.bellnet.info/?feed=rss2",
	"https://blog.tokka.shop/?xml",
	"https://web-price.info/atom.xml",
	"https://tokka1147.com/feed",
}

type FeedConfig struct {
	Selector   string
	RemoveText []string
	RemoveDiv  []string
}

var RssListsMap = map[string]FeedConfig{
	// "https://tokkataro.blog.jp/atom.xml": {
	// 	Selector:   "div.article-body",
	// 	RemoveText: []string{"●楽天市場で同じアイテムを探す", "●Yahoo!ショッピングで同じアイテムを探す", "●他の特価品を探す(ブログランキング)", "⇒激安特価！(blogranking)"},
	// },
	// "https://gekiyasu-gekiyasu.doorblog.jp/index.rdf": {
	// 	Selector:   "div.main",
	// 	RemoveText: []string{"人気ブログランキング"},
	// },
	"http://buy.livedoor.biz/index.rdf": {
		Selector:   "div.main", // div.mainが二つある
		RemoveText: []string{"適用"},
		RemoveDiv:  []string{"button"},
	},
	// "http://ks4402.blog94.fc2.com/?xml": {
	// 	Selector:   "div.entry_body",
	// 	RemoveText: []string{"(ブログランキング)", "ほかの特価情報(ブログランキング)", "関連記事", "⇒ほかの特価情報", "⇒"},
	// 	RemoveDiv:  []string{"dl.relate_dl", "u"},
	// },
}
