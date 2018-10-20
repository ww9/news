package feed

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"

	"github.com/PuerkitoBio/goquery"
	"github.com/corpix/uarand"
	"github.com/kennygrant/sanitize"
)

// Item represents a link retrieved from feed
type Item struct {
	Title string
	URL   string
	Tag   string
}

// Aggregator is the core structure than fetches feeds and saves them to html
type Aggregator struct {
	Items []Item // Ordered from newest to oldest. Always prepend new items.
	// Feeds is a map of URLs -> Titles for feeds. This needs to be stored somewhere so reader knows from where to fetch news
	Feeds map[string]string
	// We store item URLs so we know when something new appears
	KnownItems   map[string]bool
	Directory    string
	URLFetcher   func(url string) ([]byte, error)
	pages        int
	ItemsPerPage int
	NextPage     int
}

var log = logrus.New()

// SetLogger is used from main to set custom unified logger
func SetLogger(logger *logrus.Logger) {
	log = logger
}

// New creates an Aggregator with default URL fetcher
func New(directory string) (*Aggregator, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	return NewWithCustom(directory, 1000, MakeURLFetcher(client))
}

// NewWithCustom allows for creating customized Aggregators such as custom URL fetcher for testing or with custom http.client
func NewWithCustom(directory string, itemsPerPage int, URLFetcher func(URL string) ([]byte, error)) (*Aggregator, error) {
	if directory == "" {
		directory = "news"
	}
	agg := &Aggregator{
		Items:        make([]Item, 0),
		Feeds:        make(map[string]string),
		KnownItems:   make(map[string]bool),
		Directory:    filepath.Clean(directory),
		URLFetcher:   URLFetcher,
		ItemsPerPage: itemsPerPage,
		pages:        1,
	}

	if !fileExists(agg.Directory) {
		if agg.Directory == "news" {
			if errDir := os.Mkdir(agg.Directory, os.ModeDir); errDir != nil {
				return nil, fmt.Errorf("couldn't create dirextory: %s", errDir)
			}
		} else {
			return nil, fmt.Errorf("directory %s does not exist", agg.Directory)
		}
	}
	indexFile := filepath.Clean(agg.Directory + "/index.html")
	if !fileExists(indexFile) {
		if err := createSampleIndex(indexFile); err != nil {
			return nil, fmt.Errorf("could not create sample index.html file: %s", err)
		}
		log.Infof("Created %s with sample feeds.\n", indexFile)
	}

	return agg, agg.loadKnownURLs()
}

// feedXMLParser returns items ordered from oldest to newest. So we can always just append as long as template reads in inverted order.
func feedXMLParser(XML []byte) (items []Item, err error) {
	cleanXML := cleanXML(XML)
	items = make([]Item, 0)
	parser := gofeed.NewParser()
	parser.RSSTranslator = NewCustomRSSTranslator()
	feed, err := parser.ParseString(string(cleanXML))
	//feed, err := rss.Parse(cleanXML)
	if err != nil {
		return items, fmt.Errorf("could not parse XML: %s", err)
	}
	// if err != nil && strings.Contains(err.Error(), "invalid character entity") {
	// 	cleanXML = []byte(unescapeXML(string(cleanXML)))
	// 	feed, err = rss.Parse(cleanXML)
	// }
	for _, item := range feed.Items {
		itemURL := strings.TrimSpace(item.Link)
		if item.Custom["Comments"] != "" {
			itemURL = strings.TrimSpace(item.Custom["Comments"])
		}
		if itemURL == "" {
			log.Debugf("skipping item from feed %s due to lack of URL", feed.Link)
			continue
		}
		itemTitle := strings.TrimSpace(item.Title)
		if itemTitle == "" {
			itemTitle = sanitize.Name(itemURL)
			if itemTitle == "" {
				itemTitle = itemURL
			}
			log.Debugf("using %s to fill in feed %s item empty description", itemTitle, feed.Link)
		}
		items = append([]Item{Item{
			Title: itemTitle,
			URL:   itemURL,
		}}, items...)
	}
	return items, nil
}

// MakeURLFetcher is the default HTTP client used to fetch feed XML.
// The other one is fakeURLFetcher() used for testing.
// There's also a retired makeCachedURLFetcher() which was using during initial phases of development and is kept in misc.go
func MakeURLFetcher(client *http.Client) func(URL string) (content []byte, err error) {
	antiFlood := makeURLDebouncer(30 * time.Second)
	return func(URL string) (content []byte, err error) {
		req, err := http.NewRequest("GET", antiFlood(URL), nil)
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Set("User-Agent", uarand.GetRandom())
		req.Header.Set("Accept", "application/xml")
		req.Header.Set("Content-Type", "application/xml; charset=utf-8")

		resp, err := client.Do(req)
		if err != nil {
			return []byte(""), fmt.Errorf("could not open URL %s : %s", URL, err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte(""), fmt.Errorf("could not read body of URL %s : %s", URL, err)
		}
		return body, nil
	}
}

func savePageToFile(fileName string, items []Item, feeds map[string]string, nextPage int) error {
	f, err := os.Create(fileName)
	defer f.Close()
	if err != nil {
		return err
	}
	return Tpl.Execute(f, map[string]interface{}{
		"Items":    items,
		"Feeds":    feeds,
		"NextPage": nextPage,
	})
}

func loadFromFile(filePath string) (items []Item, feeds map[string]string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return items, feeds, fmt.Errorf("could not open file %s : %s", filePath, err)
	}
	defer f.Close()
	return loadFromReader(f)
}

func loadFromReader(r io.Reader) (items []Item, feeds map[string]string, err error) {
	items = make([]Item, 0)
	feeds = make(map[string]string)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return items, feeds, fmt.Errorf("could not parse HTML: %s", err)
	}
	doc.Find(".item .tag").Remove()
	doc.Find(".item").Each(func(i int, s *goquery.Selection) {
		newItem := Item{
			Title: s.Text(),
			URL:   s.AttrOr("href", ""),
		}
		newItem.SetTag()
		items = append(items, newItem)
	})
	doc.Find(".feed").Each(func(i int, s *goquery.Selection) {
		feeds[s.AttrOr("href", "")] = s.Text()
	})
	return items, feeds, nil
}

// fixRelativeURL prepends domain name to relative URLs when necessary
func (item *Item) fixRelativeURL(feedURL string) {
	if strings.HasPrefix(item.URL, "/") {
		u, err := url.Parse(feedURL)
		if err == nil {
			item.URL = u.Scheme + "://" + u.Host + item.URL
		}
	}
}

// SetTag fills .Tag based on the URL. Examples:
// /r/programming for https://www.reddit.com/r/programming/comments/9p07bh/convert_string_to_int_in_java/
// slashdot for https://science.slashdot.org/story/18/10/17/1552218/the-results-of-your-genetic-test-are-reassuring-but-that-can-change
// Hacker News for https://news.ycombinator.com/item?id=18240182
// domain.com for https://www.domain.com/item123
func (item *Item) SetTag() {
	URL := strings.ToLower(strings.TrimSpace(item.URL))
	item.Tag = ""
	if strings.Contains(URL, "reddit.com/r/") {
		parts := strings.Split(URL, "reddit.com/r/")
		parts = strings.Split(parts[1], "/")
		item.Tag = "/r/" + parts[0]
	} else if strings.Contains(URL, "slashdot.org/") {
		item.Tag = "Slashdot"
	} else if strings.Contains(URL, "news.ycombinator") {
		item.Tag = "Hacker News"
	} else if u, err := url.Parse(URL); err == nil {
		item.Tag = strings.TrimPrefix(u.Host, "www.")
	}
}

func (agg *Aggregator) loadKnownURLs() error {
	for i := 1; ; i++ {
		filePath := agg.Directory + "/index.html"
		if i > 1 {
			filePath = fmt.Sprintf(agg.Directory+"/page%d.html", i)
		}
		if !fileExists(filePath) {
			if i == 0 {
				return fmt.Errorf("could not find index.html. You need to create %s and make sure it has at least one feed URL in it. See example index.html in github", filePath)
			}
			break
		}
		agg.pages = i
		log.Debugf("reading items from %s", filePath)
		items, feeds, err := loadFromFile(filePath)
		if err != nil {
			return fmt.Errorf("could not load known URLs from file %s : %s", filePath, err)
		}
		if i == 1 {
			agg.Feeds = feeds
		}
		for _, item := range items {
			agg.KnownItems[item.URL] = true
		}
	}
	return nil
}

func createSampleIndex(file string) error {
	return savePageToFile(file, []Item{}, map[string]string{
		"https://www.reddit.com/r/golang/.rss": "/r/golang",
		"https://news.ycombinator.com/rss":     "Hacker News",
	}, 0)
}

// Update load feeds from index.html, fetches items from them and save everything back to index.html. Also generates pageX.html if necessary.
func (agg *Aggregator) Update() (err error) {
	indexFile := agg.Directory + "/index.html"
	indexItems, feeds, err := loadFromFile(indexFile)
	// If we can't read feed sources from index.html, might as well stop now
	if err != nil {
		return err
	} else if len(feeds) == 0 {
		return fmt.Errorf("zero feed sources found in file %s", indexFile)
	}
	agg.Items = indexItems
	agg.Feeds = feeds
	suffledURLs := shuffleMapKeys(agg.Feeds)
	for _, feedURL := range suffledURLs {
		log.Debugf("reading items from %s", feedURL)
		contents, err := agg.URLFetcher(feedURL)
		if err != nil {
			log.Errorf("%s : %s", feedURL, err)
			continue
		}
		items, err := feedXMLParser(contents)
		if err != nil {
			log.Errorf("%s: %s", feedURL, err)
			continue
		}
		for i := len(items) - 1; i >= 0; i-- {
			items[i].fixRelativeURL(feedURL)
			if agg.KnownItems[items[i].URL] == false {
				items[i].SetTag()
				agg.KnownItems[items[i].URL] = true
				agg.Items = append([]Item{items[i]}, agg.Items...)
			}
		}
		// Every time index.html grows too large, we shave half of its oldest items into a new page
		for len(agg.Items) >= agg.ItemsPerPage*2 {
			pageItems := agg.Items[agg.ItemsPerPage:]
			agg.pages++
			log.Debugf("saving items to page%d.html", agg.pages)
			pageFile := fmt.Sprintf(agg.Directory+"/page%d.html", agg.pages)
			if err := savePageToFile(pageFile, pageItems, agg.Feeds, agg.pages-1); err != nil {
				log.Errorf("error saving page %s : %s", pageFile, err)
				continue
			}
			agg.Items = agg.Items[:agg.ItemsPerPage]
		}
		// User might have updated feeds in index.html, so we must read it again to prevent overwriting
		_, feedsToSave, err := loadFromFile(indexFile)
		if err != nil {
			log.Errorf("error reading feeds before writing to %s: %s", indexFile, err)
			feedsToSave = agg.Feeds
		}
		if err := savePageToFile(indexFile, agg.Items, feedsToSave, agg.pages); err != nil {
			log.Errorf("error saving page %s : %s", indexFile, err)
			continue
		}
	}
	return nil
}
