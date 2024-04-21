package scraper

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id string
	title string
	company string
	// location string
}

func Scrape(term string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?&searchword=" + term
	var jobs []extractedJob
	totalPages := getPages(baseURL)
	c := make(chan []extractedJob)

	for i := 0; i < totalPages; i++ {
		go getPage(baseURL, i + 1, c)
	}

	for j := 0; j < totalPages; j++ {
		extractJobs := <-c
		jobs = append(jobs, extractJobs...)
	}

	writeJobs(jobs)
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Company"}
	
	err = w.Write(headers)
	checkErr(err)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.company}
		err = w.Write(jobSlice)
		checkErr(err)
	}
}

func getPage(baseURL string, page int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)

	pageURL := baseURL + "&recruitPage=" + strconv.Itoa(page)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	items := doc.Find(".item_recruit")
	items.Each(func(i int, item *goquery.Selection) {
		go extractJob(item, c)
	})

	for i := 0; i < items.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(job *goquery.Selection, c chan<- extractedJob) {
	id, _ := job.Attr("value")
	title := CleanString(job.Find(".job_tit>a").Text())
	company := CleanString(job.Find(".corp_name>a").Text())
	c <- extractedJob {
		id: id,
		title: title,
		company: company}
	// location := s.Find(".job_condition>a").Text()
	// fmt.Println(location)
}

func getPages(baseURL string) int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	
	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with status:", res.StatusCode)
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}