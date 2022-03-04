package main

import (
	"fmt"
	"github.com/smallnest/ringbuffer"

)

func main() {
	rb := ringbuffer.New(1024)
	fmt.Println(rb)
}