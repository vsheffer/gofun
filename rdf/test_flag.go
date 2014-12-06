package main

import (
	"flag"
	"fmt"
)

func main() {
	var str = flag.String("strarg", "default", "This is a string arg.")
	flag.Parse()
	fmt.Printf("str = %s", *str)
}
