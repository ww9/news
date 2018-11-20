package feed

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/sirupsen/logrus"
)

// Tests if new items are read and written in the correct order to the HTML file. This isn't fully automated
// One should run this test then open test_data/news to check if it has correct feeds and items
func Test_HTMLFileGeneratedOnNewItems(t *testing.T) {
	dir := "test_data/news"
	failIfError(t, os.RemoveAll(dir))
	failIfError(t, os.MkdirAll(dir, os.ModeDir))
	failIfError(t, copyFileContents("test_data/1feed_0items.html", dir+"/index.html"))
	agg, err := NewWithCustom(logrus.New(), dir, 4, fakeURLFetcher)
	failIfError(t, err)
	if len(agg.Feeds) != 1 {
		t.Errorf("Expected to have 1 feed imported from test_data/1feed_0items.html but found %d", len(agg.Feeds))
	}
	if agg.Feeds["https://www.reddit.com/r/golang/.rss"] != "/r/golang" {
		t.Error("Could not find expected '/r/golang' feed")
	}
	// It's safe to call Update() 40 times in hsort succession since we are using a fake URL fectcher
	for i := 0; i < 40; i++ {
		failIfError(t, agg.Update())
	}

}

// Tests happy path when importing OPML file including ovewriting of pre-existent feed URL
func Test_OPMLFileImport(t *testing.T) {
	dir := "test_data/news"
	failIfError(t, os.RemoveAll(dir))
	failIfError(t, os.MkdirAll(dir, os.ModeDir))
	failIfError(t, copyFileContents("test_data/1feed_0items.html", dir+"/index.html"))
	agg, err := NewWithCustom(logrus.New(), dir, 1000, fakeURLFetcher)
	failIfError(t, err)
	failIfError(t, agg.ImportOPMLFile("test_data/feeds.opml"))
	// we know the imported OPML file has 69 items including "https://www.reddit.com/r/golang/.rss" which
	// should overwrite the one feed loaded from 1feed_0items.html making our feed URL count be 69
	if len(agg.Feeds) != 69 {
		t.Errorf("Expected 69 feeds URL in aggregator but found %d", len(agg.Feeds))
	}
	//
	if agg.Feeds["https://www.reddit.com/r/golang/.rss"] != "/r/golang/ (test: title ovewritten from OPML)" {
		t.Error("Title of feed with URL \"https://www.reddit.com/r/golang/.rss\" wasn't ovewritten from OPML")
	}
	failIfError(t, agg.Update())
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

// toRSSXML is used for testing. It returns an RSS XML string containing a randomly
// picked feed and all items from the provided Aggregator instance.
func toRSSXML(agg *Aggregator) string {
	feedTitle := ""
	feedURL := ""
	// Get random feed since we are iterating a map
	for feedURL, feedTitle = range agg.Feeds {
		break
	}
	xml := `<rss version="2.0">
	<channel>
	<title>` + feedTitle + `</title>
	<link>` + feedURL + "</link>\n"

	for _, item := range agg.Items {
		xml += "<item><title>" + item.Title + "</title><link>" + item.URL + "</link></item>\n"
	}
	return xml + "</channel>\n</rss>"
}

// Maybe Go 2.0 can help us with this verbosity :)
func failIfError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
