package main

import (
	"io/ioutil",
	"fmt"
)

func main() {
	fileInfo, _ := ioutil.ReadDir("/tmp/specfiles")
	for file := range fileInfo {
		fmt.Printf("file = %+v", file)
	}
}