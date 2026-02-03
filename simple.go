package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
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
	defer resp.Body.Close()

	var result itmePricesPrices
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalln(err)
	}

	for _, item := range result.Data.Items {
		fmt.Printf("Item: %s | Avg Price: %d\n", item.ShortName, item.Avg24hPrice)
	}
}

type graphqlRequest struct {
	Query string `json:"query"`
}

type itmePricesPrices struct {
	Data struct {
		Items []struct {
			Id           string `json:"id"`
			ShortName    string `json:"shortName"`
			Avg24hPrice  int    `json:"avg24hPrice"`
			High24hPrice int    `json:"high24hPrice"`
			Low24hPrice  int    `json:"low24hPrice"`
			IconLink     string `json:"iconLink"`
			// FIXME: Bitcoin's price is 0
			// TODO: Explore the sellFor field to fix ^
		} `json:"items"`
	} `json:"data"`
}
