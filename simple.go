package main

import (
	"fmt"
	"io" 
	"log"
	"net/http"
	"strings"
)

func main()  {
	body := strings.NewReader(`{"query": "{ items {id name Moonshine } }"}`)
	req, err := http.NewRequest("POST", "https://api.tarkov.dev/graphql", body)
	if err != nil {
	    log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	bodyString := string(bodyBytes)
	fmt.Println(bodyString)

	defer resp.Body.Close()
}
