package main

import (
	"flag"
	"fmt"
)

var n = flag.Bool("n", false, "omit trailing newline")
var sep = flag.String("s", " ", "separator")
var x = flag.Int("i", 0, "hhh")

func main() {
	fmt.Println("hello")
	flag.Parse()
	fmt.Println(*n, *sep, *x)
	// p1:=f()
	// p2:=f()
	// if(p1==p2){
	// 	fmt.Println("equal")
	// }
	// fmt.Println(p1,p2)
	// fmt.Println(*p1)
}
func f() *int {
	v := 123
	return &v
}
