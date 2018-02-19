package main

import (
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Book struct {
	Title            string   `json:"title"`
	Author           string   `json:"author"`
	LibrarianComment string   `json:"librarianComment"`
	Tags             []string `json:"tags"`
	Content          []string `json:"content"`
	Source           string   `json:"source"`
}

func newBookFromHTMLNode(doc *html.Node, source string) Book {
	book := Book{
		Source: source,
	}

	if titleNode := htmlquery.FindOne(doc, `//*[@id="main"]/h1`); titleNode != nil {
		book.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
	}

	if authorNode := htmlquery.FindOne(
		doc, `//*[@class="node node-book"]/div[3]/div[1]/div/div`,
	); authorNode != nil {
		text := strings.Replace(
			htmlquery.InnerText(authorNode),
			"Author:", "", 1,
		)
		book.Author = strings.TrimSpace(text)
	}

	if commentNode := htmlquery.FindOne(
		doc, `//*[@class="node node-book"]/div[3]/div[2]/div/div/p`,
	); commentNode != nil {
		text := htmlquery.InnerText(commentNode)
		book.LibrarianComment = strings.TrimSpace(text)
	}

	for _, n := range htmlquery.Find(
		doc,
		`//*[@class="node node-book"]/div[2]/ul/li/a`,
	) {
		text := strings.TrimSpace(htmlquery.InnerText(n))
		book.Tags = append(book.Tags, text)
	}

	for _, n := range htmlquery.Find(
		doc,
		`//*[@class="node node-book"]/div[3]/p`,
	) {
		text := strings.TrimSpace(htmlquery.InnerText(n))
		book.Content = append(book.Content, text)
	}

	return book
}

func (b Book) Print() {
	fmt.Printf(
		`%% %s
%% %s
%% %s
%% %s
%% %s
`,
		b.Title,
		b.Author,
		b.LibrarianComment,
		b.Tags,
		b.Source,
	)
	fmt.Println()

	for _, paragraph := range b.Content {
		fmt.Println(paragraph)
		fmt.Println()
	}
}
