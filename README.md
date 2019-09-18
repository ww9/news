[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/) [![Go Report Card](https://goreportcard.com/badge/github.com/ww9/news)](https://goreportcard.com/report/github.com/ww9/news)

# 📰 News
News is a minimalist RSS/Atom aggregator that saves to HTML files.
```
📂news
  ├📰 index.html
  ├📰 page2.html
  └📰 page3.html
```

That's it! No database, no configuration files, no HTTP server, no ads, no tracking and no JavaScript. Everything is stored in the HTML files which look like this:

![screenshot](screenshot.png)

## Usage

Running `news` creates `📂news` directory containing a sample `📰index.html` file which you should edit with your own RSS/Atom feed sources.

Every 10 minutes it fetches news from your feeds and saves what's fresh to `📰index.html`.

When `📰index.html` grows large (1000 items by default), the oldest 500 items are moved to `📰page2.html`.

`📂news` can reside in Google Drive or Dropbox for easy access everywhere. 

This is how I use it:

```bash
news -wait 30 -dir "D:/gdrive/news"
```

## Command-line arguments
`news -h` prints:
```
  -dir string
        directory to store html files. By default ./news is used and created if necessary
  -items int
        number of items per page.html file. A new page.html file is created whenever index.html contains 2x that number (default 500)
  -noflood int
        minium seconds between calls to same domain to avoid flooding (default 30)
  -opml string
        path to OPML file containing feed URLS to be imported. Existing feed URLs are ovewritten, not duplicated
  -template news/feed/template.go
        custom Go html/template file to use when generating .html files. See news/feed/template.go
  -timeout int
        timeout in seconds when fetching feeds (default 10)
  -verbose
        verbose mode outputs extra info when enabled
  -wait int
        minutes to wait between updates (default 10)
```

## Running from code
`go get -uv github.com/ww9/news`

`cd $GOROOT/src/github.com/ww9/news`

`go get ./...` to fetch dependencies

`go run main.go`

## Installing from code
`go install -i github.com/ww9/news`

If you have Go's `/bin` directory in `$PATH` env variable, you should be able to run `news` from anywhere.

## Downloading binaries
Windows, Linux and OSX binaries are available in [Releases](https://github.com/ww9/news/releases).

## Todo

- [ ] More tests
- [ ] Go modules
- [ ] Vendor
- [ ] Dockerfile

## License

[The Unlicense](http://unlicense.org/), [Public Domain](https://gist.github.com/ww9/4c4481fb7b55186960a34266078c88b1). As free as it gets.