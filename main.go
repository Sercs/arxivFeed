package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/telluz/gotex"
)

func main() {

	// Key variables
	docName := time.Now().Local().Format("02-01-2006") // save as?
	savePath := ""                                     // save where?

	// query what?
	// see https://arxiv.org/help/api/user-manual for options
	// Note: You'd have to change the URL below for different query types
	queries := [...]string{"cat:q-bio.NC", "cat:cs.CV", "cat:cs.LG", "cat:cs.AI", "cat:cs.SI"}
	maxResults := "25" // query how many?

	// Intialise LaTeX Document
	var document = `
	\documentclass{article}
	\usepackage{hyperref}
	\usepackage[utf8]{inputenc}
	\usepackage[T1]{fontenc}
	\title{Article Feed}
	\begin{document}
	\maketitle
	\tableofcontents
	\newpage
	`

	// Loop through all queries
	for _, query := range queries {

		// Construct a query
		URL := "http://export.arxiv.org/api/query?max_results=" + maxResults + ";search_query=" + query + "&sortBy=submittedDate&sortOrder=descending"
		fp := gofeed.NewParser()
		resp, err := fp.ParseURL(URL)
		if err != nil {
			log.Fatal(err)
		}

		document = document + `\section{` + query + `}`
		document = formatLatex(resp, document)
		time.Sleep(30 * time.Second) // Don't steal all the bandwidth!
	}

	document = document + `\end{document}`

	// Compile the document
	pdf, err := gotex.Render(document, gotex.Options{Command: "", Runs: 3})
	if err != nil {
		log.Println("Render Failed: ", err)
	}

	// Create a file to store the PDF
	file, err := os.Create(savePath + docName + ".pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Write it
	file.Write(pdf)
}

func formatLatex(resp *gofeed.Feed, document string) string {

	// Loop through all returned articles
	for i, article := range resp.Items {
		fmt.Println("Article: ", i)
		fmt.Println(article.Title)
		// filter for today and remove articles which I haven't figured out how to format (i.e. $'s)
		if lazyRemove(article) {
			fmt.Println("Pass")
			// Fancy formatting
			document = document + `\subsection{` + cleanForLatex(article.Title) + `}`
			document = document + `{\scriptsize \textit{Published: ` + cleanDate(article.Published)
			document = document + `}}\\`
			document = document + cleanForLatex(article.Description)
			authors := ""
			for i, author := range article.Authors {
				if i < len(article.Authors)-1 {
					authors = authors + author.Name + ", "
				} else {
					authors = authors + author.Name
				}
			}

			document = document + `\begin{center}`
			document = document + `\scriptsize ` + article.Link
			document = document + `\end{center}`

			document = document + `\begin{flushright}`
			document = document + `\textbf{\footnotesize ` + authors + `}`
			document = document + `\end{flushright}`
			document = document + `\normalsize`
		} else {
			fmt.Println("Fail")
		}
	}
	return document
}

// I think this would require some search between $...$ function and if so, don't remove stuff
func lazyRemove(article *gofeed.Item) bool {
	if !strings.Contains(article.Description, "$") {
		return true
	} else {
		return false
	}
}

// Filter for today's content (technically yesterday because time-zones)
func filterDate(article *gofeed.Item) bool {
	if cleanDate(article.Updated) == time.Now().Local().AddDate(0, 0, -1).Format("2006-01-02") {
		return true
	} else {
		return false
	}
}

// Clean abstracts for LaTeX format
func cleanForLatex(s string) string {
	s = strings.ReplaceAll(s, "\\%", "")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "&", "\\&")
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "^", "\\^")
	s = strings.ReplaceAll(s, "~", "$\\sim$")
	s = strings.ReplaceAll(s, "\\textcolor{black}", "")
	s = strings.ReplaceAll(s, "\\mu", "$\\mu$")
	s = strings.ReplaceAll(s, "\\alpha", "$\\alpha$")
	s = strings.ReplaceAll(s, "\\cite", "")

	return s
}

func cleanDate(s string) string {
	spl := strings.Split(s, "T")
	return spl[0]
}
