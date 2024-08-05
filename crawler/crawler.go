package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	opts    *CrawlerOpts
	results map[string]struct{}

	wg *sync.WaitGroup
	mu sync.RWMutex

	targetURL    string
	targetDomain string
}

type CrawlerOpts struct {
	TargetURL   string
	IsRecursive bool
}

func New(opts *CrawlerOpts) (*Crawler, error) {
	u, err := url.Parse(opts.TargetURL)
	if err != nil {
		return nil, err
	}

	return &Crawler{
		opts:         opts,
		results:      make(map[string]struct{}),
		targetURL:    opts.TargetURL,
		targetDomain: u.Host,
		mu:           sync.RWMutex{},
		wg:           &sync.WaitGroup{},
	}, nil
}

// urls parse document, get all urls, filter it
func (c *Crawler) urls(targetURL string) ([]string, error) {
	var urls []string

	resp, err := http.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL err: %v; url: %s", err, targetURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("error fetching URL: status code %d; url: %s", resp.StatusCode, targetURL)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading HTML: %v", err)
	}

	doc.Find("a").Each(func(index int, item *goquery.Selection) {
		href, exists := item.Attr("href")
		if exists {
			parsedURL, err := url.Parse(href)
			if err == nil {

				if !parsedURL.IsAbs() {
					baseURL, err := url.Parse(targetURL)
					if err == nil {
						href = baseURL.ResolveReference(parsedURL).String()
					}
				}
				urls = append(urls, href)

			}
		}
	})

	urls = c.filter(urls)

	return urls, nil

}

// Results returns slice of URLs
func (c *Crawler) Results() []string {
	results := make([]string, 0, len(c.results))
	for l := range c.results {
		results = append(results, l)
	}
	return results
}

// filter matches incoming URLs with target domain name, also skips all the same links but with anchors
func (c *Crawler) filter(urls []string) []string {
	var filteredURLs []string
	var parsedURL *url.URL
	var cleanedURL string
	var err error

	uniqueURLs := make(map[string]bool)

	for _, u := range urls {
		parsedURL, err = url.Parse(u)
		if err != nil {
			continue
		}

		if parsedURL.Host == c.targetDomain {
			cleanedURL = strings.Split(u, "#")[0]

			if !uniqueURLs[cleanedURL] {
				uniqueURLs[cleanedURL] = true

				if c.checkUrl(cleanedURL) {
					continue
				}

				filteredURLs = append(filteredURLs, cleanedURL)
			}
		}
	}

	return filteredURLs
}

// checkUrl thread safe check link in map
func (c *Crawler) checkUrl(url string) bool {
	c.mu.RLock()
	_, ok := c.results[url]
	c.mu.RUnlock()

	return ok
}

// checkAndAdd - thread safe check and add link to map
func (c *Crawler) checkAndAdd(url string) bool {
	c.mu.Lock()
	_, ok := c.results[url]
	if !ok {
		c.results[url] = struct{}{}
	}
	c.mu.Unlock()

	return ok
}
