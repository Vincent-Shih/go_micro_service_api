// pure golang client
package main

import (
	"bufio"
	"net/http"
)

func main() {
	// plain get is able to stream the data
	res, err := http.Get("http://localhost:8080/sse")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)

	// keep streaming out
	for scanner.Scan() {
		println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
