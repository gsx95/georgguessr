package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"
)
import "fmt"
import "io/ioutil"

func main() {
	tries := 1
	available := false
	fmt.Println("Waiting for local setup to be available....")
	for tries <= 10 && !available {
		fmt.Printf("Try to reach local setup: %d\n", tries)
		available = tryReachLocalSetup()
		if !available {
			tries++
			time.Sleep(30 * time.Second)
		}
	}
	if !available {
		panic(fmt.Sprintf("Setup not reachable after try %d", tries))
	}

	fmt.Println("Setup is reachable, starting tests")
}

func tryReachLocalSetup() bool {

	_, err := net.DialTimeout("tcp","127.0.0.1:3000", 30 * time.Second)
	if err != nil {
		fmt.Println(err)
		return false
	}
	response, err := http.Get("http://127.0.0.1:3000/exists/{roomID}")
	if err != nil {
		fmt.Println(err)
		return false
	}
	if response.StatusCode != 404 {
		fmt.Printf("Status code: %d, Expecting 404\n", response.StatusCode)
		return false
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	type resp struct {
		exists bool
	}
	var jsonBody resp
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		fmt.Println("error parsing body: " + string(body))
		fmt.Println(err)
		return false
	}
	return true
}
