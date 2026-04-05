package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall/js"
)

func main() {
	// Use a channel to keep the program alive
	keepAlive := make(chan struct{})

	// Register the function
	// The first argument "getTarkovPrices" MUST match window.getTarkovPrices()
	js.Global().Set("getTarkovPrices", js.FuncOf(func(this js.Value, args []js.Value) any {
		return runAsPromise(func() (any, error) {
			res, err := getPrices()
			return res.Data.Items, err
		})
	}))
	js.Global().Set("getTarkovRecipes", js.FuncOf(func(this js.Value, args []js.Value) any {
		return runAsPromise(func() (any, error) {
			res, err := getRecipes()
			return res.Data.Items, err
		})
	}))

	fmt.Println("Go functions registered!")
	<-keepAlive
}

func sendQuery(query string, target interface{}) error {
	payload := graphqlRequest{Query: query}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "https://api.tarkov.dev/graphql", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func runAsPromise(fetchFunc func() (any, error)) js.Value {
	handler := js.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		reject := args[1]

		go func() {
			data, err := fetchFunc()
			if err != nil {
				reject.Invoke(js.ValueOf(err.Error()))
				return
			}

			jsonBytes, _ := json.Marshal(data)
			resolve.Invoke(js.ValueOf(string(jsonBytes)))
		}()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}

// TODO: handle error if the api is down
func getPrices() (itemPrices, error) {
	var result itemPrices
	err := sendQuery(pricesString, &result)
	if err != nil {
		return result, err
	}

	for i := range result.Data.Items {
		maxPrice := 0
		for _, vendorPrices := range result.Data.Items[i].SellFor {
			if vendorPrices.PricesRUB > maxPrice {
				maxPrice = vendorPrices.PricesRUB
			}
		}
		if result.Data.Items[i].LastLowPrice > maxPrice {
			maxPrice = result.Data.Items[i].LastLowPrice
		}
		result.Data.Items[i].BestPrice = maxPrice
	}

	return result, nil
}

func getRecipes() (itemRecipes, error) {
	var result itemRecipes
	err := sendQuery(recipesString, &result)
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
			BestPrice    int    `json:"bestPrice"` // Best selling price
			LastLowPrice int    `json:"lastLowPrice"`
			SellFor      []struct {
				PricesRUB int `json:"priceRUB"`
			} `json:"sellFor"`
		} `json:"items"`
	} `json:"data"`
}

const pricesString = `
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
				"544fb6cc4bdc2d34748b456e"
				"567143bf4bdc2d1a0f8b4567"
				"5d1b371186f774253763a656"
				"5d1b2fa286f77425227d1674"
				"5c06779c86f77426e00dd782"
				"5d1b309586f77425227d1676"
				"5d03775b86f774203e7e0c4b"
			]
			gameMode: pve
		) {
			shortName
			iconLink
			id
			lastLowPrice
			sellFor {
				priceRUB
			}
		}
	}
	`

type itemRecipes struct {
	Data struct {
		Items []struct {
			Name       string `json:"name"`
			ShortName  string `json:"shortName"`
			IconLink   string `json:"iconLink"`
			Id         string `json:"id"`
			BartersFor []struct {
				RequiredItems []struct {
					Quantity int `json:"quantity"`
					Item     struct {
						ShortName string `json:"shortName"`
						IconLink  string `json:"iconLink"`
						Id        string `json:"id"`
					} `json:"item"`
				} `json:"requiredItems"`
			} `json:"bartersFor"`
		} `json:"items"`
	} `json:"data"`
}

const recipesString = `	
	{
		items(
			lang: en
			ids: [
				"5b6d9ce188a4501afc1b2b25"
				"5c0a840b86f7742ffa4f2482"
				"59fb023c86f7746d0d4b423c"
			]
			gameMode: pve
		) {
			name
			shortName
			iconLink
			bartersFor {
				requiredItems {
					item {
						shortName
						iconLink
						id
					}
					quantity
				}
			}
		}
	}
	`
