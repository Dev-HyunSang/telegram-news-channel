package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/dev-hyunsang/telegram-news-channel/config"
)

// https://app.quicktype.io/
type News struct {
	Status       string    `json:"status"`
	TotalResults int64     `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type Article struct {
	Source      Source      `json:"source"`
	Author      string      `json:"author"`
	Title       string      `json:"title"`
	Description interface{} `json:"description"`
	URL         string      `json:"url"`
	URLToImage  interface{} `json:"urlToImage"`
	PublishedAt string      `json:"publishedAt"`
	Content     interface{} `json:"content"`
}

type Source struct {
	ID   ID   `json:"id"`
	Name Name `json:"name"`
}

type ID string

const (
	GoogleNews ID = "google-news"
)

type Name string

const (
	NameGoogleNews Name = "Google News"
)

// HTTP RESPONSE to JSON
func UnmarshalNews(data []byte) (News, error) {
	var r News
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *News) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func GetBreakingNewsHeadline() {
	resp, err := http.Get(
		fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=kr&apiKey=%s", config.GetEnv("NEWS_API_KEY")))
	if err != nil {
		log.Println("[ERROR] BreakingNewsHeadline | Failed Request News API")
		panic(err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("[ERROR] BreakingNewsHeadline | Failed Read Response Body")
		panic(err)
	}

	result, err := UnmarshalNews(data)
	if err != nil {
		log.Println("[ERROR] BreakingNewsHeadline | Failed Unmarshal Response Body")
		panic(err)
	}

	// log.Println(result.Articles[1])
	// =>  {{google-news Google News} 한겨레 구속영장 기각 뒤 '또 필로폰 투약' 남경필 장남 구속 - 한겨레 <nil> https://news.google.com/rss/articles/CBMiNmh0dHBzOi8vd3d3LmhhbmkuY28ua3IvYXJ0aS9hcmVhL3llb25nbmFtLzEwODYxMDMuaHRtbNIBAA?oc=5 <nil> 2023-04-01T07:56:18Z <nil>}

	for _, newsData := range result.Articles {
		// SendChannel(newsData.Title + newsData.URL)
		SendMultipartData(newsData.Title, newsData.URL)
	}
}

func TopHeadlineNews() {
}

// https://core.telegram.org/bots/api#sendmessage
func SendChannel(text string) {
	resp, err := http.Get(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s", config.GetEnv("TELEGRAM_API_KEY"), config.GetEnv("TELGRAM_CHANNEL_ID"), text))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	log.Println(string(data))
}

func SendMultipartData(title, url string) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	text := fmt.Sprintf(`<a href="%s"><b>%s</b></a>`, url, title)
	_ = writer.WriteField("chat_id", config.GetEnv("TELGRAM_CHANNEL_ID"))
	_ = writer.WriteField("text", text)
	_ = writer.WriteField("parse_mode", "HTML")
	_ = writer.WriteField("disable_web_page_preview", "false")

	err := writer.Close()
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?", config.GetEnv("TELEGRAM_API_KEY")), payload)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))
}

func main() {
	GetBreakingNewsHeadline()
}
