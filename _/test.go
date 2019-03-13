package main

import (
	"fmt"
	"strconv"
)

func main() {
	v := 3.14159265355436365476879679780780786755646453

	s32 := strconv.FormatFloat(v, 'g', +10, 32)
	fmt.Printf("%T, %v\n", s32, s32)

	s64 := strconv.FormatFloat(v, 'f', -1, 64)
	fmt.Printf("%T, %v\n", s64, s64)

}
