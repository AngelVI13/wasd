package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/gen2brain/beeep"
)

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

var checksum string

var args struct {
	URL   string `arg:"-u,--url" default:"https://www.coutellerie-tourangelle.com/sp_nouveautes.php" help:"URL to check for changes."`
	Sleep int    `arg:"-s,--sleep" default:"90" help:"Sleep time between checks."`
}

type Item struct {
	Name string
	ID   int
}

func websiteItemsBody(url string) (string, error) {
	log.Printf("Checking %s", url)
	resp, err := http.Get(url)
	checkError(err)

	b, err := ioutil.ReadAll(resp.Body)
	checkError(err)

	body := string(b)

	start := `<script type="application/ld+json">`
	end := "</script>"
	startIdx := strings.Index(body, start)
	endIdx := strings.LastIndex(body, end)

	if startIdx == -1 || endIdx == -1 {
		return "", fmt.Errorf(
			"couldn't find website data between\n%s\nand\n%s\n",
			start,
			end,
		)
	}

	return body[startIdx+len(start) : endIdx], nil
}

func processItems(jsonBody string) map[int]Item {
	var rawItems any

	err := json.Unmarshal([]byte(jsonBody), &rawItems)
	checkError(err)

	items, ok := rawItems.([]any)
	if !ok {
		panic("incorrect type")
	}

	allItems := map[int]Item{}
	for _, elem := range items {
		item, ok := elem.(map[string]any)
		if !ok {
			log.Fatalf("unexpected type for element")
		}

		name, found := item["name"]
		if !found {
			continue
		}

		rawId, found := item["productID"]
		if !found {
			continue
		}

		id := int(rawId.(float64))

		allItems[id] = Item{
			Name: name.(string),
			ID:   id,
		}
	}

	return allItems
}

func compareItems(old map[int]Item, new map[int]Item) []string {
	var newItems []string

	for k, v := range new {
		_, found := old[k]
		if !found {
			newItems = append(newItems, v.Name)
		}
	}

	return newItems
}

var currentItems map[int]Item

func main() {
	arg.MustParse(&args)
	url := args.URL

	for {
		jsonBody, err := websiteItemsBody(url)
		checkError(err)

		items := processItems(jsonBody)

		newItems := compareItems(currentItems, items)

		if len(newItems) > 0 && len(currentItems) > 0 {
			currentItems = items
			newItemsText := strings.Join(newItems, ", ")
			log.Println("Found new items: ", newItemsText)
			log.Println("Sending notification")
			err := beeep.Notify(
				"New item/s was added!",
				fmt.Sprintf("New arrival in the shop: %s", newItemsText),
				"assets/information.png",
			)
			checkError(err)
		} else if len(currentItems) == 0 {
			log.Println("Got initial values")
			currentItems = items
		} else if len(newItems) == 0 {
			log.Println("No new items")
		}

		log.Printf("Sleeping for %d seconds", args.Sleep)
		time.Sleep(time.Duration(args.Sleep) * time.Second)
	}
}
