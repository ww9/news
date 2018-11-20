package feed

import (
	"fmt"

	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/rss"
)

// CustomRSSTranslator is currently used to instruct mmcdole/gofeed to parse <comments> as item link in Reddit's feed
type CustomRSSTranslator struct {
	defaultTranslator *gofeed.DefaultRSSTranslator
}

// NewCustomRSSTranslator creates a new XML feed translator to be used by our instance of mmcdole/gofeed
func NewCustomRSSTranslator() *CustomRSSTranslator {
	t := &CustomRSSTranslator{}
	t.defaultTranslator = &gofeed.DefaultRSSTranslator{}
	return t
}

// Translate is called by mmcdole/gofeed
func (ct *CustomRSSTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
	rss, found := feed.(*rss.Feed)
	if !found {
		return nil, fmt.Errorf("Feed did not match expected type of *rss.Feed")
	}
	f, err := ct.defaultTranslator.Translate(rss)
	if err != nil {
		return nil, err
	}
	for i, item := range rss.Items {
		f.Items[i].Custom = map[string]string{"Comments": item.Comments}
	}
	return f, nil
}

// If we ever need an Atom translator here's the skeleton:

// type CustomAtomTranslator struct {
// 	defaultTranslator *gofeed.DefaultAtomTranslator
// }

// func NewCustomAtomTranslator() *CustomAtomTranslator {
// 	t := &CustomAtomTranslator{}
// 	t.defaultTranslator = &gofeed.DefaultAtomTranslator{}
// 	return t
// }

// func (ct *CustomAtomTranslator) Translate(feed interface{}) (*gofeed.Feed, error) {
// 	atom, found := feed.(*atom.Feed)
// 	if !found {
// 		return nil, fmt.Errorf("Feed did not match expected type of *atom.Feed")
// 	}
// 	f, err := ct.defaultTranslator.Translate(atom)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for i, item := range atom.Entries {
// 		f.Items[i].Custom = map[string]string{"Comments": item.Comments}
// 	}
// 	return f, nil
// }
