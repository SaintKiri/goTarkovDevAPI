package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"syscall/js"
)

func main() {
	// Use a channel to keep the program alive
	keepAlive := make(chan struct{})

	// Register the function
	// The first argument "getTarkovPrices" MUST match window.getTarkovPrices()
	js.Global().Set("getTarkovPrices", js.FuncOf(getPricesWrapper))

	fmt.Println("Go functions registered!")
	<-keepAlive
}

func getPricesWrapper(this js.Value, args []js.Value) any {
	// We use a Promise-based approach because networking is asynchronous
	handler := js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		reject := args[1]

		// Run the logic in a goroutine so we don't block the main thread
		go func() {
			result, err := getPrices()
			if err != nil {
				reject.Invoke(js.ValueOf(err.Error()))
				return
			}

			// Convert the Go struct to a JSON string to pass back to JS
			// Alternatively, you could build a js.Value object manually
			jsonBytes, _ := json.Marshal(result.Data.Items)
			resolve.Invoke(js.ValueOf(string(jsonBytes)))
		}()

		return nil
	})

	promiseConstructor := js.Global().Get("Promise")
	return promiseConstructor.New(handler)
}

// func printPrices() {
// 	var result itemPrices = getPrices()
// 	for _, item := range result.Data.Items {
// 		var price int = item.Avg24hPrice
// 		if price == 0 {
// 			for _, vendorPrices := range item.SellFor {
// 				if vendorPrices.PricesRUB > price {
// 					price = vendorPrices.PricesRUB
// 				}
// 			}
// 		}
// 		fmt.Printf("Item: %s | Price: %d\n", item.ShortName, price)
// 	}
// }

func getPrices() (itemPrices, error) {
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
			shortName
			iconLink
			id
			avg24hPrice
			high24hPrice
			low24hPrice
			sellFor {
				priceRUB
			}
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

	var result itemPrices
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalln(err)
	}

	return result, err
}

type graphqlRequest struct {
	Query string `json:"query"`
}

type itemPrices struct {
	Data struct {
		Items []struct {
			ShortName    string `json:"shortName"`
			IconLink     string `json:"iconLink"`
			Id           string `json:"id"`
			Avg24hPrice  int    `json:"avg24hPrice"`
			High24hPrice int    `json:"high24hPrice"`
			Low24hPrice  int    `json:"low24hPrice"`
			SellFor      []struct {
				PricesRUB int `json:"priceRUB"`
			} `json:"sellFor"`
		} `json:"items"`
	} `json:"data"`
}
