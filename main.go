package main

import (
	"crypto"
	"io/ioutil"
	"log"
	"net/http"
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

func main() {
	arg.MustParse(&args)
	url := args.URL

	for {
		log.Printf("Checking %s", url)
		resp, err := http.Get(url)
		checkError(err)

		b, err := ioutil.ReadAll(resp.Body)
		checkError(err)

		hash := crypto.SHA256.New()
		_, err = hash.Write(b)
		checkError(err)

		newChecksum := string(hash.Sum(nil))
		log.Printf("Hash %x", hash.Sum(nil))

		if newChecksum != checksum && checksum != "" {
			checksum = newChecksum
			log.Println("Sending notification")
			err := beeep.Notify(
				"New item was added!",
				"New arrival in the shop",
				"assets/information.png",
			)
			checkError(err)
		} else if checksum == "" {
			checksum = newChecksum
		}

		log.Printf("Sleeping for %d seconds", args.Sleep)
		time.Sleep(time.Duration(args.Sleep) * time.Second)
	}
}
