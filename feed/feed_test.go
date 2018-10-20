package feed

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"
)

// Tests if new items are read and written in the correct order to the HTML file
// This isn't fully automated, one should run this test then open /rss/test_data/news
func Test_HTMLFileGeneratedOnNewItems(t *testing.T) {
	dir := "test_data/news"
	failIfError(t, os.RemoveAll(dir))
	os.MkdirAll(dir, os.ModeDir)
	failIfError(t, copyFileContents("test_data/index_empty.html", dir+"/index.html"))
	r, err := NewWithCustom(dir, 4, fakeURLFetcher)
	failIfError(t, err)
	for i := 0; i < 40; i++ {
		failIfError(t, r.Update())
	}
}

var fakeFeedItemID = int64(0)

// fakeURLFetcher generates fake items with incrementing IDs.
// It always return 2 old items and 2 new items
func fakeURLFetcher(URL string) (content []byte, err error) {
	agg := &Aggregator{
		Feeds: map[string]string{URL: "Title of " + URL},
		Items: make([]Item, 0),
	}
	atomic.AddInt64(&fakeFeedItemID, 2)
	for i := fakeFeedItemID - 2; i < fakeFeedItemID+2; i++ {
		agg.Items = append(agg.Items, Item{
			Title: fmt.Sprintf("Item %d Title", i),
			URL:   fmt.Sprintf("%s?item=%d", URL, i),
		})
	}
	return []byte(toRSSXML(agg)), nil
}

// toRSSXML is used to generate fake feeds XML for testing. It returns RSS XML containing one feed and all items from an Aggregator.
func toRSSXML(agg *Aggregator) string {
	feedTitle := ""
	feedURL := ""
	// Get random feed
	for feedURL, feedTitle = range agg.Feeds {
		break
	}
	xml := `<rss version="2.0">
<channel>
<title>` + feedTitle + `</title>
<link>` + feedURL + `</link>
`
	for _, item := range agg.Items {
		xml += `<item><title>` + item.Title + `</title><link>` + item.URL + `</link></item>
`
	}
	return xml + `</channel>
</rss>`
}

// Maybe Go 2.0 can help us with this verbosity :)
func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
