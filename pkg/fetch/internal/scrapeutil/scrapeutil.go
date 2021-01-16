package scrapeutil

import (
	"regexp"

	"go.uber.org/zap"
)

var reHref = regexp.MustCompile(`link\s*=\s*['"](https?://[^'"]+)['"]`)
var reMarkdownLink = regexp.MustCompile(`\]\((https?://[^\)]+)\)`)

func ScrapeLinks(hypertext string) []string {
	// zap.S().Debugf("hypertext: %s", hypertext)

	hrefMatches := reHref.FindAllStringSubmatch(hypertext, -1)
	mdMatches := reMarkdownLink.FindAllStringSubmatch(hypertext, -1)

	links := make([]string, 0, len(hrefMatches)+len(mdMatches))
	for _, match := range hrefMatches {
		link := match[1]
		zap.S().Debugf("Scraped off href: %s", link)
		links = append(links, link)
	}
	for _, match := range mdMatches {
		link := match[1]
		zap.S().Debugf("Scraped off mdLink: %s", link)
		links = append(links, link)
	}
	return links
}
