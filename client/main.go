package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

var DEFAULT_RBDNS_ENDPOINT = "http://localhost:8080"

var ADD_RECORD string = "addRecord"
var QUERY string = "query"
var EXIT string = "exit"

func main() {
	// Single invoke mode
	if len(os.Args) == 4 {
		if os.Args[1] != ADD_RECORD {
			fmt.Println("Usage: rbdns-client addRecord {key} {value}")
			os.Exit(0)
		}
		startTime := time.Now()
		resp := addRecord(os.Args[2], os.Args[3])
		endTime := time.Now()
		fmt.Println(resp)
		// Originally in nanoseconds. Divide by 1 million to get milliseconds
		elapsed := float64(endTime.Sub(startTime)) / float64(1000000)
		fmt.Printf("%vms\n", elapsed)
		return
	}

	if len(os.Args) == 3 {
		if os.Args[1] != QUERY {
			fmt.Println("Usage: rbdns-client query {key}")
			os.Exit(0)
		}
		startTime := time.Now()
		resp := query(os.Args[2])
		endTime := time.Now()
		fmt.Println(resp)
		// Originally in nanoseconds. Divide by 1 million to get milliseconds
		elapsed := float64(endTime.Sub(startTime)) / float64(1000000)
		fmt.Printf("%vms\n", elapsed)
		return
	}

	// REPL mode
	fmt.Println("RBDNS server: start")
	fmt.Println("available commands:")
	fmt.Println("addRecord {key} {value}")
	fmt.Println("query {key}")
	fmt.Println("exit")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("> ")
	input := read(reader)
	for ; input[0] != EXIT; input = read(reader) {
		switch getCommand(input) {

		case ADD_RECORD:
			if len(input) != 3 {
				fmt.Println("Usage: addRecord {key} {value}")
			}

			startTime := time.Now()
			resp := addRecord(input[1], input[2])
			endTime := time.Now()
			fmt.Println(resp)
			// Originally in nanoseconds. Divide by 1 million to get milliseconds
			elapsed := float64(endTime.Sub(startTime)) / float64(1000000)
			fmt.Printf("%vms\n", elapsed)

		case QUERY:
			if len(input) != 2 {
				fmt.Println("Usage: query {key}")
			}

			startTime := time.Now()
			resp := query(input[1])
			endTime := time.Now()
			fmt.Println(resp)
			elapsed := float64(endTime.Sub(startTime)) / float64(1000000)
			fmt.Printf("%vms\n", elapsed)

		default:
			fmt.Println("Unrecognized command")

		}
		fmt.Print("> ")
	}

	fmt.Println("RBDNS client: closing. Bye!")
}

func addRecord(key string, value string) (okMsg string) {
	rpc := fmt.Sprintf("/addRecord?key=%s&value=%s", key, value)
	resp, err := http.Get(DEFAULT_RBDNS_ENDPOINT + rpc)
	if err != nil {
		fmt.Println("GET /addRecord failed")
		return "not OK"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("addRecord() error in client app. This should not have happened.")
		return "not OK"
	}

	return string(body)
}

func query(key string) (value string) {
	rpc := fmt.Sprintf("/query?key=%s", key)
	resp, err := http.Get(DEFAULT_RBDNS_ENDPOINT + rpc)
	if err != nil {
		fmt.Println("GET /query failed")
		return "not OK"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("query() error in client app. This should not have happened.")
		return "not OK"
	}

	return string(body)
}

func read(r *bufio.Reader) []string {
	t, _ := r.ReadString('\n')
	return strings.Split(strings.TrimSpace(t), " ")
}

func getCommand(input []string) string {
	return input[0]
}
