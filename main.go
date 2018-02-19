package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
		"arena":       "/books/arena/by-title",
		"daggerfall":  "/books/daggerfall/by-title",
		"battlespire": "/books/battlespire/by-title",
		"redguard":    "/books/redguard/by-title",
		"morrowind":   "/books/morrowind/by-title",
		"shadowkey":   "/books/shadowkey/by-title",
		"oblivion":    "/books/oblivion/by-title",
		"skyrim":      "/books/skyrim/by-title",
		"online":      "/books/online/by-title",
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

		err = traverseBooks(*flagURL, func(doc *html.Node, url string) error {
			book := newBookFromHTMLNode(doc, url)
			books = append(books, book)

			return nil
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

		resp, err := http.Get("https://" + filepath.Join(rootURL, url))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			return err
		}

		nodes := htmlquery.Find(doc,
			`//*[@id="content"]/div/div[2]/div/ul/li/span/span/a`,
		)

		for _, n := range nodes {
			err := traverseBooks("https://"+filepath.Join(rootURL, n.Attr[0].Val),
				func(doc *html.Node, url string) error {
					node := htmlquery.FindOne(
						doc,
						`//*[@id="main"]`,
					)
					if node == nil {
						return fmt.Errorf(
							"node main not found in doc: %#v", doc,
						)
					}

					path := filepath.Join("imperial-library", game)
					if err := os.MkdirAll(path, os.ModePerm); err != nil {
						return err
					}

					filename := url[strings.LastIndex(url, "/"):] + ".html"

					file, err := os.Create(filepath.Join(path, filename))
					if err != nil {
						return err
					}

					fmt.Println("rendering", game, filename)

					if err := html.Render(file, node); err != nil {
						return err
					}

					if err := file.Close(); err != nil {
						return err
					}

					return nil
				},
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func traverseBooks(url string, fn func(*html.Node, string) error) error {
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
		if err := fn(doc, url); err != nil {
			return err
		}
	}

	return nil
}

func findSerieLinks(doc *html.Node) []string {
	var links []string
	for _, n := range htmlquery.Find(
		doc, `//*[@class="book-navigation"]/ul[@class="menu"]/li/a`,
	) {
		links = append(links, "https://"+filepath.Join(rootURL, n.Attr[0].Val))
	}

	if len(links) == 0 {
		return nil
	}
	return links
}

func makeEbook() error {
	return nil
}
