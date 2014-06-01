package main

import "github.com/segmentio/go-loggly-search"
import . "github.com/bitly/go-simplejson"
import "fmt"
import "os"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	c := search.New(os.Getenv("ACCOUNT"), os.Getenv("USER"), os.Getenv("PASS"))

	res, err := c.Query(`(login OR logout) AND tobi`).Size(50).From("-5h").Fetch()
	check(err)

	for _, event := range res.Events {
		Output(event)
	}
	fmt.Println()
}

func Output(event interface{}) {
	msg := event.(map[string]interface{})["logmsg"].(string)
	obj, err := NewJson([]byte(msg))
	check(err)

	fmt.Println()
	for k, v := range obj.MustMap() {
		fmt.Printf("  %14s: %s\n", k, v)
	}
}
