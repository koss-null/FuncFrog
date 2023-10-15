package main

import (
	"errors"
	"fmt"

	"github.com/koss-null/funcfrog/pkg/pipe"
)

func main() {
	p := pipe.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	y := pipe.NewYeti()
	res := p.Yeti(y).Map(func(i int) int {
		if i < 0 {
			y.Yeet(errors.New("omg the value is negative"))
		}
		return 2 * i
	}).Snag(func(err error) {
		fmt.Println("Snagging an error: " + err.Error())
	}).Do()
	fmt.Println(res)
}
