package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ww9/news/feed"
)

var flagDir = flag.String("d, dir", "", "directory to store html files. By default ./news is used and created if necessary")
var flagTimeout = flag.Int("timeout", 10, "timeout in seconds when fetching feeds")
var flagUpdateInterval = flag.Int("wait", 10, "minutes to wait between updates")
var flagItemsPerPage = flag.Int("items", 500, "number of items per page.html file. A new page.html file is created whenever index.html contains 2x that number")
var flagVerbose = flag.Bool("v, verbose", false, "verbose mode outputs extra info when enabled")
var flagTemplateFile = flag.String("template", "", "custom Go html/template file to use when generating .html files. See `news/feed/template.go`")
var flagOPMLFile = flag.String("opml", "", "path to OPML file containing feed URLS to be imported. Existing feed URLs are ovewritten, not duplicated")
var flagMinDomainRequestInterval = flag.Int("noflood", 30, "minium seconds between calls to same domain to avoid flooding")

func main() {
	flag.Parse()
	*flagTimeout = minMax(*flagTimeout, 1, 60)
	*flagItemsPerPage = minMax(*flagItemsPerPage, 2, 500)
	*flagUpdateInterval = minMax(*flagUpdateInterval, 1, 24*60)
	*flagMinDomainRequestInterval = minMax(*flagMinDomainRequestInterval, 10, 24*60*60)

	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	if *flagVerbose {
		log.SetLevel(logrus.DebugLevel)
	}

	if *flagTemplateFile != "" {
		tpl, err := template.ParseFiles(*flagTemplateFile)
		if err != nil {
			log.Fatalf("Could not load custom template file: %s", err)
		}
		feed.Tpl = tpl
	}
	agg, err := feed.NewWithCustom(
		log,
		*flagDir,
		*flagItemsPerPage,
		feed.MakeURLFetcher(
			log,
			time.Second*time.Duration(*flagMinDomainRequestInterval),
			&http.Client{Timeout: time.Second * time.Duration(*flagTimeout)},
		),
	)
	if err != nil {
		log.Fatalln(err)
	}
	if *flagOPMLFile != "" {
		importedFeeds, err := agg.ImportOPMLFile(*flagOPMLFile)
		if err != nil {
			log.Fatalf("Could not import OPML file: %s", err)
		} else {
			log.Printf("Successfully imported %d feeds from OPML file.", importedFeeds)
		}
	}

	go func() {
		for {
			log.Infof("Fetching news from %d feed sources...", len(agg.Feeds))
			if err := agg.Update(); err != nil {
				log.Fatalln(err)
			}
			log.Infof("Done. Waiting %d minutes for next update...", *flagUpdateInterval)
			time.Sleep(time.Duration(*flagUpdateInterval) * time.Minute)
		}
	}()

	pressCTRLCToExit()
	fmt.Println("Bye :)")
}

func pressCTRLCToExit() {
	exitCh := make(chan os.Signal)
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		exitCh <- (<-signalCh)
	}()
	<-exitCh
}

func minMax(value int, min int, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}
