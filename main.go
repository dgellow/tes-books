package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	rootURL = "www.imperial-library.info"
	// libraryURL = "https://www.imperial-library.info/books/all/by-title"
)

var (
	flagCommand = flag.String("command", "",
		"Command to run: download, bookpage (requires -url)")
	flagURL   = flag.String("url", "", "")
	flagGames = flag.String("games", "",
		"Games to restrict the command to: arena, daggerfall, battlespire, redguard, morrowind, shadowkey, oblivion, skyrim, online")

	bookLists = map[string]string{
		"arena":       "https://www.imperial-library.info/books/arena/by-title",
		"daggerfall":  rootURL + "/books/daggerfall/by-title",
		"battlespire": rootURL + "/books/battlespire/by-title",
		"redguard":    rootURL + "/books/redguard/by-title",
		"morrowind":   rootURL + "/books/morrowind/by-title",
		"shadowkey":   rootURL + "/books/shadowkey/by-title",
		"oblivion":    rootURL + "/books/oblivion/by-title",
		"skyrim":      rootURL + "/books/skyrim/by-title",
		"online":      rootURL + "/books/online/by-title",
	}
)

func main() {
	flag.Parse()

	var err error
	switch *flagCommand {
	case "download":
		games := strings.Split(*flagGames, ",")
		err = download(games, bookLists)
	case "bookpage":
		if *flagURL == "" {
			flag.Usage()
			os.Exit(-1)
		}

		var books []Book

		err = traverseBooks(*flagURL, func(doc *html.Node, url string) {
			book := newBookFromHTMLNode(doc, url)
			books = append(books, book)
		})

		b, err := json.Marshal(books)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(b))
	default:
		flag.Usage()
		os.Exit(-1)
	}

	if err != nil {
		panic(err)
	}
}

func download(selectedGames []string, sources map[string]string) error {
	for game, url := range sources {
		found := false
		for _, g := range selectedGames {
			if g == game {
				found = true
				break
			}
		}
		if !found {
			continue
		}

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
