package main

import (
	"net/http"
	"os"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/typester/go-pit"
)

var titleRe = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)

func getTitle(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	matches := titleRe.FindStringSubmatch(string(content))
	if matches == nil {
		return "", fmt.Errorf("can't retrieve page title\n")
	}

	return matches[1], nil
}

func postedJSON(key string, url string) (string, error) {
	type Request struct {
		Key     string `json:"key"`
		LongURL string `json:"longUrl"`
	}

	req := Request{
		Key:     key,
		LongURL: url,
	}

	bytes, err := json.Marshal(&req)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

const apiURL = "https://www.googleapis.com/urlshortener/v1/url"

type shortenURLResponse struct {
	ID string `json:"id"`
}

func decodeAPIResponse(r io.Reader) (string, error) {
	decoder := json.NewDecoder(r)

	var res shortenURLResponse
	if err := decoder.Decode(&res); err != nil {
		return "", err
	}

	return res.ID, nil
}

func shortURL(url string) (string, error) {
	profile, err := pit.Get("goo.gl", pit.Requires{
		"key": "API key",
	})

	if err != nil {
		return "", err
	}

	jsonStr, err := postedJSON((*profile)["key"], url)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	return decodeAPIResponse(res.Body)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: url_title url\n")
		return
	}

	url := os.Args[1]

	title, err := getTitle(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	shorten, err := shortURL(url)
	fmt.Printf("%s %s", title, shorten)
}
