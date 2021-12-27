package main

//TODO Concurrency
//TODO CI tests

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Person struct {
	url        string
	name       string
	species    string
	department string
	state      string
	eow        time.Time
	cause      string
}

func getCopsPerYear(year string) []Person {

	yearURL := fmt.Sprintf("https://www.odmp.org/search/year/%v", year)
	res, err := http.Get(yearURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var copArr []Person
	doc.Find(".officer-short-details").Each(func(_ int, s *goquery.Selection) {
		var name string
		var species string
		var url string
		var department string
		var state string
		var eow string
		var eowDate time.Time
		var cause string
		var causeClean string
		s.Find("p").Each(func(i int, p *goquery.Selection) {
			if i == 0 {
				name = p.Text()
				species = parseDog(name)
				url, _ = p.Find("a").Attr("href")
			} else if i == 1 {
				department = p.Text()
				department, state = cleanDepartment(department)

			} else if i == 2 {
				eow = p.Text()
				eowDate = convertEOW(eow)
			} else if i == 3 {
				var cop Person
				cause = p.Text()
				causeClean = cleanCause(cause)
				cop = Person{
					url:        url,
					name:       name,
					species:    species,
					department: department,
					state:      state,
					eow:        eowDate,
					cause:      causeClean,
				}
				copArr = append(copArr, cop)
			}
		})
	})
	return copArr
}

func convertEOW(p string) time.Time {
	eow := p[5:]
	date, _ := time.Parse("Monday, January 2, 2006", eow)
	return date
}

func cleanCause(p string) string {
	eow := p[7:]
	return eow
}

func parseDog(p string) string {
	k, _ := regexp.Compile("K9")
	b := []byte(p)
	t := k.Match(b)
	if t {
		return "dog"
	} else {
		return "human"
	}
}

func cleanDepartment(p string) (string, string) {
	eow := strings.Split(p, ",")
	return eow[0], eow[1][1:]
}
func writeOutToCSV(year string, data []Person) {
	fiName := fmt.Sprintf("%v.csv", year)
	file, _ := os.Create(fiName)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	header := []string{"name", "species", "cause", "department", "state", "url"}

	writer.Write(header)
	for _, value := range data {
		datLine := []string{value.name, value.species, value.cause, value.department, value.state, value.url}
		writer.Write(datLine)
	}
}

func main() {
	year := "2020"
	cops := getCopsPerYear(year)
	writeOutToCSV(year, cops)

}
