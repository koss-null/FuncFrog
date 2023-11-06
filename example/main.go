package main

import (
	"errors"
	"fmt"

	"github.com/koss-null/funcfrog/pkg/pipe"
)

func main() {
	p := pipe.Slice([]int{1, 2, 3, 4, -5, 6, 7, 8, 9})
	y1, y2 := pipe.NewYeti(), pipe.NewYeti()
	res := p.Yeti(y1).Map(func(i int) int {
		if i < 0 {
			y1.Yeet(errors.New("omg the value is NEGATIVE"))
		}
		return i - 6
	}).Snag(func(err error) {
		fmt.Println("Snagging an error: " + err.Error())
	}).Yeti(y2).Map(func(i int) int {
		if i > 0 {
			y2.Yeet(errors.New("omg the value is POSITIVE"))
		}
		return 2 * i
	}).Snag(func(err error) {
		fmt.Println("another snag for the same error: " + err.Error())
	}).Filter(func(i *int) bool { return *i > 0 }).Do()
	fmt.Println(res)
}
