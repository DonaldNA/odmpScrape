package main

//TODO Concurrency
//TODO CI tests

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PersonBio struct {
	url           string
	age           string
	tour          string
	veteran       bool
	incidentDate  time.Time
	officerWeapon string
	offenderLabel string
	lat           string
	long          string
}

// func getCopsPerYear(year string) []Person {

// 	yearURL := fmt.Sprintf("https://www.odmp.org/search/year/%v", year)
// 	res, err := http.Get(yearURL)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer res.Body.Close()
// 	if res.StatusCode != 200 {
// 		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
// 	}

// 	// Load the HTML document
// 	doc, err := goquery.NewDocumentFromReader(res.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	var copArr []Person
// 	doc.Find(".officer-short-details").Each(func(_ int, s *goquery.Selection) {
// 		var name string
// 		var species string
// 		var url string
// 		var department string
// 		var state string
// 		var eow string
// 		var eowDate time.Time
// 		var cause string
// 		var causeClean string
// 		s.Find("p").Each(func(i int, p *goquery.Selection) {
// 			if i == 0 {
// 				name = p.Text()
// 				species = parseDog(name)
// 				url, _ = p.Find("a").Attr("href")
// 			} else if i == 1 {
// 				department = p.Text()
// 				department, state = cleanDepartment(department)

// 			} else if i == 2 {
// 				eow = p.Text()
// 				eowDate = convertEOW(eow)
// 			} else if i == 3 {
// 				var cop Person
// 				cause = p.Text()
// 				causeClean = cleanCause(cause)
// 				cop = Person{
// 					url:        url,
// 					name:       name,
// 					species:    species,
// 					department: department,
// 					state:      state,
// 					eow:        eowDate,
// 					cause:      causeClean,
// 				}
// 				copArr = append(copArr, cop)
// 			}
// 		})
// 	})
// 	return copArr
// }

// func writeOutToCSV(year string, data []Person) {
// 	fiName := fmt.Sprintf("%v.csv", year)
// 	file, _ := os.Create(fiName)
// 	defer file.Close()

// 	writer := csv.NewWriter(file)
// 	defer writer.Flush()
// 	header := []string{"name", "species", "cause", "department", "state", "url"}

// 	writer.Write(header)
// 	for _, value := range data {
// 		datLine := []string{value.name, value.species, value.cause, value.department, value.state, value.url}
// 		writer.Write(datLine)
// 	}
// }

func getCOPsFromCSV(data [][]string) []string {
	var urlArr []string
	for i, line := range data {
		if i > 0 { // omit header line
			for j, field := range line {
				if j == 5 {
					urlArr = append(urlArr, field)
				}
			}
		}
	}
	return urlArr
}

// Custom user agent.
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

// -------------------------------------

// fetchUrl opens a url with GET method and sets a custom user agent.
// If url cannot be opened, then log it to a dedicated channel.
func fetchUrl(url string, chFailedUrls chan string, chIsFinished chan bool) {

	// Open url.
	// Need to use http.Client in order to set a custom user agent:
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)

	// Inform the channel chIsFinished that url fetching is done (no
	// matter whether successful or not). Defer triggers only once
	// we leave fetchUrl():
	defer func() {
		chIsFinished <- true
	}()

	// If url could not be opened, we inform the channel chFailedUrls:
	if err != nil || resp.StatusCode != 200 {
		chFailedUrls <- url
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// 	var copArr []Person
	// test := doc.Find(".officer-bio li").Text()
	// test := doc.Find(".incident-details li").Text()
	// test := doc.Find("#officer-map-data").Text() //Coords
	fmt.Println(test)

	// 		var namincident-detailse string
	// 		var species string
	// 		s.Find("p").Each(func(i int, p *goquery.Selection) {

}

func main() {
	// open file
	f, err := os.Open("2020.csv")
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	copArr := getCOPsFromCSV(data)
	// Create 2 channels, 1 to track urls we could not open
	// and 1 to inform url fetching is done:
	chFailedUrls := make(chan string)
	chIsFinished := make(chan bool)

	// Open all urls concurrently using the 'go' keyword:
	for _, url := range copArr {
		go fetchUrl(url, chFailedUrls, chIsFinished)
	}

	// Receive messages from every concurrent goroutine. If
	// an url fails, we log it to failedUrls array:
	failedUrls := make([]string, 0)
	for i := 0; i < len(copArr); {
		select {
		case url := <-chFailedUrls:
			failedUrls = append(failedUrls, url)
		case <-chIsFinished:
			i++
		}
	}

	// Print all urls we could not open:
	fmt.Println("Could not fetch these urls: ", failedUrls)

}
