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

var flagDir = flag.String("d", "", "directory to save html files in. If none specified ./news is used and created if necessary")
var flagTimeout = flag.Int("t", 10, "timeout in seconds when fetching feeds")
var flagUpdateInterval = flag.Int("i", 10, "minutes to wait between updates")
var flagItemsPerPage = flag.Int("n", 500, "number of items per page.html file. A new page.html file is created whenever index.html contains 2x that number")
var flagVerbose = flag.Bool("v", false, "verbose mode outputs extra info when enabled")
var flagTemplateFile = flag.String("c", "", "custom Go html/template file to to use when generating .html files. See `news/feed/template.go` in source for an example")

func main() {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	feed.SetLogger(log)

	flag.Parse()
	*flagTimeout = minMax(*flagTimeout, 1, 60)
	*flagItemsPerPage = minMax(*flagItemsPerPage, 2, 500)
	*flagUpdateInterval = minMax(*flagUpdateInterval, 1, 24*60)
	if *flagVerbose {
		log.SetLevel(logrus.DebugLevel)
	}
	if *flagTemplateFile != "" {
		tpl, err := template.ParseFiles(*flagTemplateFile)
		if err != nil {
			log.Fatalf("Could not load custom template file: %s\n", err)
		}
		feed.Tpl = tpl
	}

	go func() {
		agg, err := feed.NewWithCustom(*flagDir, *flagItemsPerPage, feed.MakeURLFetcher(&http.Client{Timeout: time.Second * time.Duration(*flagTimeout)}))
		if err != nil {
			log.Fatalln(err)
		}
		for {
			log.Infof("Fetching news from %d feed sources...\n", len(agg.Feeds))
			if err := agg.Update(); err != nil {
				log.Fatalln(err)
			}
			log.Infof("Done. Waiting %d minutes for next update...\n", *flagUpdateInterval)
			time.Sleep(time.Duration(*flagUpdateInterval) * time.Minute)
		}
	}()

	pressCTRLCToExit()
}

func pressCTRLCToExit() {
	fmt.Printf("\nPress CTRL+C to exit\n")
	exitCh := make(chan struct{})
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		select {
		case <-signalCh:
		}
		exitCh <- struct{}{}
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
