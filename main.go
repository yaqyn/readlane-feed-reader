package main

import (
	"fmt"
	"github.com/yaqyn/blog-aggregator/internal/config"
	"os"
)

func main() {
	c, e := config.Read()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	if e = c.SetUser("yaqyn"); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	c, e = config.Read()
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", c)
}
