package feed

import (
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"

	"github.com/kennygrant/sanitize"
)

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func fileExists(f string) bool {
	_, err := os.Stat(f)
	return !os.IsNotExist(err)
}

func makeURLDebouncer(log *logrus.Logger, wait time.Duration) func(URL string) string {
	lastAccessed := make(map[string]time.Time)
	return func(URL string) string {
		u, err := url.Parse(URL)
		if err != nil {
			log.Errorf("Could not parse URL %s, will try to fetch it anyway: %s", URL, err)
			return URL
		}
		domain := u.Host
		timePassed := time.Since(lastAccessed[domain])
		if timePassed < wait {
			log.Debugf("Waiting %.1f seconds to request from %s", (wait - timePassed).Seconds(), URL)
			time.Sleep(wait - timePassed)
		}
		lastAccessed[domain] = time.Now()
		return URL
	}
}

func cleanXML(XML []byte) []byte {
	// http://blog.zikes.me/post/cleaning-xml-files-before-unmarshaling-in-go/
	// Remove utf8 special shenanigans
	printOnly := func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}
	return []byte(strings.Map(printOnly, string(XML)))
}

// Fixes "could not parse HTML: XML syntax error on line 1: invalid character entity &mdash;" in codes like this:
// <summary type="html">Every site I&#039;ve used Codeship on&mdash;until today...</summary>
// See: https://stackoverflow.com/questions/3805050/xml-parser-error-entity-not-defined
func unescapeXML(XML string) string {
	return html.UnescapeString(XML)
}

// makeCachedURLFetcher is retired. It was used during initial phases of development to prevent spamming feed sources.
func makeCachedURLFetcher(log *logrus.Logger, minDomainRequestInterval time.Duration, client *http.Client, cacheDir string) func(URL string) (content []byte, err error) {
	fetcher := MakeURLFetcher(log, minDomainRequestInterval, client)
	return func(URL string) (content []byte, err error) {
		cacheDir = filepath.Clean(cacheDir)
		fileName := cacheDir + "/" + sanitize.BaseName(URL) + ".html"
		if fileExists(fileName) {
			log.Debugf("Cached %s", fileName)
			b, err := ioutil.ReadFile(fileName)
			if err != nil {
				return []byte(""), fmt.Errorf("error opening cache %s : %s", URL, err)
			}
			return b, nil
		}
		log.Debugf("Web %s", URL)
		content, err = fetcher(URL)
		if err != nil {
			return content, err
		}

		f, err := os.Create(fileName)
		if err != nil {
			return nil, fmt.Errorf("could not open cache file file %s for writing: %s", fileName, err)
		}
		defer f.Close()
		if _, err = f.Write(content); err != nil {
			return []byte(""), fmt.Errorf("could not write to cache file %s: %s", fileName, err)
		}
		return content, nil
	}
}

func shuffleMapKeys(srcMap map[string]string) (mapKeys []string) {
	mapKeys = make([]string, 0, len(srcMap))
	for k := range srcMap {
		mapKeys = append(mapKeys, k)
	}
	rand.Shuffle(len(mapKeys), func(i, j int) {
		mapKeys[i], mapKeys[j] = mapKeys[j], mapKeys[i]
	})
	return mapKeys
}
