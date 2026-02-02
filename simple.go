package main

import (
	"fmt"
	"io" 
	"log"
	"net/http"
	// "strings"
	"encoding/json"
	"bytes"
)

func main()  {
	queryString := `
	{
		items(
			lang: en
			ids: [
			"5d1b376e86f774252519444e"
			"5d40407c86f774318526545a"
			"5d403f9186f7743cac3f229b"

			"5c052e6986f7746b207bc3c9"
			"5c0530ee86f774697952d952"
			"5af0548586f7743a532b7e99"
			"57347c93245977448d35f6e3"

			"6389c8c5dbfd5e4b95197e6b"
			"61bf7c024770ee6f9c6b8b53"
			"590c621186f774138d11ea29"


			"59faff1d86f7746c51718c9c"
			"5d235a5986f77443f6329bc6"

			"59fb023c86f7746d0d4b423c"
			"5aafbde786f774389d0cbc0f"
			"61bf7b6302b3924be92fa8c3"
			]
			gameMode: pve
		) {
			id
			shortName
			avg24hPrice
			high24hPrice
			low24hPrice
			iconLink
		}
	}
	`
	payload := graphqlRequest{Query: queryString}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.tarkov.dev/graphql", bytes.NewBuffer(body))
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

type graphqlRequest struct {
	Query string `json:"query"`
}

type itmePricesPrices struct {
	data struct {
		items []struct {
			id string `json:"id"`
			shortName string `json:"shortName"`
			avg24hPrice int `json:"avg24hPrice"`
			high24hPrice int `json:"high24hPrice"`
			low24hPrice int `json:"low24hPrice"`
			iconLink string `json:"iconLink"`
		} `json:"items"`
	} `json:"data"`
}
