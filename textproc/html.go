package textproc

import (
	"bytes"
	"fmt"
	"net/url"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/daominah/gomicrokit/log"
	"golang.org/x/net/html"
)

// HtmlXpath finds all html nodes match the xpath query
func HtmlXpath(root *html.Node, xpath0 string) (
	nodes []*html.Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			nodes = nil
			err = fmt.Errorf("exception :\n%v", string(debug.Stack()))
		}
	}()

	_, err = xpath.Compile(xpath0)
	if err != nil {
		return nil, err
	}
	nodes = htmlquery.Find(root, xpath0)
	return nodes, nil
}

// HtmlGetText is slow, caller should reuse the result if possible.
func HtmlGetText(node *html.Node) string {
	excludedTags := map[string]bool{"script": true, "style": true}

	var buf bytes.Buffer
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			isExcluded := false
			if n.Parent != nil {
				isExcluded = excludedTags[n.Parent.Data]
			}
			if !isExcluded {
				buf.WriteString(n.Data)
				buf.WriteString("\n")
			}
		}
		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}

	f(node)
	result := buf.String()
	result = strings.TrimSpace(result)
	result = RemoveRedundantSpace(result)
	result = NormalizeText(result)

	return result
}

// HtmlGetHrefs returns all url in the html as absolute url,
// urls with different fragments are treated as one url
func HtmlGetHrefs(baseUrlStr string, node *html.Node) []string {
	result := make([]string, 0)
	setUrls := make(map[string]bool)

	baseUrl, parseBaseErr := url.Parse(baseUrlStr)
	if parseBaseErr != nil {
		log.Infof("error url.Parse %v: %v", baseUrl, parseBaseErr)
	}
	elems, _ := HtmlXpath(node, "//a/@href")
	for _, elem := range elems {
		if elem.FirstChild != nil {
			relativeUrlStr := elem.FirstChild.Data
			url0, _ := url.Parse(relativeUrlStr)
			if parseBaseErr == nil && baseUrl != nil {
				url0 = baseUrl.ResolveReference(url0)
			}
			url0.Fragment = ""
			setUrls[url0.String()] = true
		}
	}

	for k, _ := range setUrls {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// HtmlGetImgSrc returns absolute url of the image
func HtmlGetImgSrc(baseUrlStr string, imgNode *html.Node) string {
	var imgSrcS string
	for _, attr := range imgNode.Attr {
		if attr.Key == "src" {
			imgSrcS = attr.Val
		}
	}

	// convert relative url to absolute url
	baseUrl, _ := url.Parse(baseUrlStr)
	imgSrc, _ := url.Parse(imgSrcS)
	if baseUrl != nil {
		imgSrc = baseUrl.ResolveReference(imgSrc)
	}

	return imgSrc.String()
}
