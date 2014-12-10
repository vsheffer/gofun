// Simple package to support providing more then one string argument to the
// a command line program.
//
// To use with the flag package:
//
//     package main
//
//     import (
//         "github.com/vsheffer/gofun/util"
//         "flag"
//     )
//
//     func main() {
//         var stringVals util.StringSlice
//         flag.Val(&stringVals, "stringVals", nil, "This will be a list of stringVals.")
//         for index, stringVal range stringVals {
//             fmt.Printf("stringVal[%d] = %s\n", index, stringVal)
//         }
//     }
package util

import (
	"fmt"
)

type StringSlice []string

func (s *StringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}

func (s *StringSlice) Get() []string {
	return []string(*s)
}

// The second method is Set(value string) error
func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
