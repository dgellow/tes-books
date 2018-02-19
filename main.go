package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	rootURL = "www.imperial-library.info"
	// libraryURL = "https://www.imperial-library.info/books/all/by-title"
	// arenaURL       = "https://www.imperial-library.info/books/arena/by-title"
	daggerfallURL  = rootURL + "/books/daggerfall/by-title"
	battlespireURL = rootURL + "/books/battlespire/by-title"
	redguardURL    = rootURL + "/books/redguard/by-title"
	morrowindURL   = rootURL + "/books/morrowind/by-title"
	shadowkeyURL   = rootURL + "/books/shadowkey/by-title"
	oblivionURL    = rootURL + "/books/oblivion/by-title"
	skyrimURL      = rootURL + "/books/skyrim/by-title"
	onlineURL      = rootURL + "/books/online/by-title"
)

var (
	command    = flag.String("command", "", "Command to run: download, convert, bookpage + url")
	commandURL = flag.String("url", "", "")
)

func main() {
	flag.Parse()

	var err error
	switch *command {
	case "download":
		err = download()
	case "bookpage":
		if *commandURL == "" {
			flag.Usage()
			os.Exit(-1)
		}

		var books []Book

		err = traverseBooks(*commandURL, func(doc *html.Node, url string) {
			book := newBookFromHTMLNode(doc, url)
			books = append(books, book)
		})

		b, err := json.Marshal(books)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(b))
	case "convert":
		err = convert()
	default:
		flag.Usage()
		os.Exit(-1)
	}

	if err != nil {
		panic(err)
	}
}

func download() error {
	for _, url := range []string{
		skyrimURL,
	} {
		resp, err := http.Get("https://" + url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			return err
		}

		var books []Book

		for _, n := range htmlquery.Find(
			doc,
			`//*[@id="content"]/div/div[2]/div/ul/li/span/span/a`,
		) {
			err := traverseBooks("https://"+path.Join(rootURL, n.Attr[0].Val),
				func(doc *html.Node, url string) {
					book := newBookFromHTMLNode(doc, url)
					books = append(books, book)
				},
			)
			if err != nil {
				return err
			}
		}

		b, err := json.Marshal(books)
		if err != nil {
			return err
		}

		fmt.Println(b)
	}

	return nil
}

func traverseBooks(url string, fn func(*html.Node, string)) error {
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		return err
	}

	serieLinks := findSerieLinks(doc)
	if serieLinks != nil {
		for _, link := range serieLinks {
			err := traverseBooks(link, fn)
			if err != nil {
				return err
			}
		}
	} else {
		fn(doc, url)
	}

	return nil
}

func findSerieLinks(doc *html.Node) []string {
	var links []string
	for _, n := range htmlquery.Find(
		doc, `//*[@class="book-navigation"]/ul[@class="menu"]/li/a`,
	) {
		links = append(links, "https://"+path.Join(rootURL, n.Attr[0].Val))
	}

	if len(links) == 0 {
		return nil
	}
	return links
}

func makeEbook() error {
	return nil
}
