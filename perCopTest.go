package main

//TODO Concurrency
//TODO CI tests

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PersonBio struct {
	url string
	// age string
	// tour          string
	veteran bool
	// incidentDate  time.Time
	// officerWeapon string
	// offenderLabel string
	// lat           string
	// long          string
}

func writeOutToCSVTwo(year string, data []PersonBio) {
	fiName := fmt.Sprintf("%v.csv", year)
	file, _ := os.Create(fiName)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	header := []string{"url", "vet"}

	writer.Write(header)
	for _, value := range data {
		datLine := []string{value.url, strconv.FormatBool(value.veteran)}
		writer.Write(datLine)
	}
}

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
func fetchUrlTwo(i int, url string, chFailedUrls chan string, chIsFinished chan bool, copChan chan PersonBio) {

	// Open url.
	// Need to use http.Client in order to set a custom user agent:
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	fmt.Println(i)

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
	var vetStat bool
	vet := doc.Find(".officer-bio li strong").Text()
	if vet == "Military Veteran" {
		vetStat = true
	} else {
		vetStat = false
	}
	cop := PersonBio{url: url, veteran: vetStat}
	copChan <- cop
	time.Sleep(5 * time.Second)
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
	concurrencyLimit := 2
	copChan := make(chan PersonBio, concurrencyLimit)

	// Open all urls concurrently using the 'go' keyword:
	for i, url := range copArr {
		go fetchUrlTwo(i, url, chFailedUrls, chIsFinished, copChan)

	}

	// Receive messages from every concurrent goroutine. If
	// an url fails, we log it to failedUrls array:

	failedUrls := make([]string, 0)
	copBios := make([]PersonBio, 0)
	for i := 0; i < len(copArr); {
		c := <-copChan
		copBios = append(copBios, c)
		select {
		case url := <-chFailedUrls:
			failedUrls = append(failedUrls, url)
		case <-chIsFinished:
			i++
		}
	}

	writeOutToCSVTwo("2020_out", copBios)

	// Print all urls we could not open:
	fmt.Println("Could not fetch these urls: ", failedUrls)

}
