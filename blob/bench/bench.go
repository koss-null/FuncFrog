package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/koss-null/funcfrog/pkg/ff"
	"github.com/pkg/profile"
)

type User struct {
	ID string
}

func fib(i int) int {
	if i == 0 || i == 1 {
		return 1
	}
	i--
	prev := 1
	cur := 1
	for i > 0 {
		cur, prev = prev+cur, cur
		i--
	}
	return cur
}

func GetUserID(u *User) string {
	return strconv.Itoa(fib(int(u.ID[0]) + 10))
}

func main() {
	makeUsers(1)
	defer profile.Start().Stop()
	// n == number of users
	start := time.Now()
	for _, n := range []int{
		1,
		100,
		10_000,
		1_000_000,
		10_000_000,
	} {
		users := makeUsers(n)
		for i := 0; i < 3; i++ {
			_ = ff.Map(users, GetUserID).Parallel(4).Do()
		}
	}
	fmt.Println("done in", time.Since(start))

	start = time.Now()
	for _, n := range []int{
		1,
		100,
		10_000,
		1_000_000,
		10_000_000,
	} {
		users := makeUsers(n)
		for i := 0; i < 3; i++ {
			res := make([]string, 0, len(users))
			for j := range users {
				res = append(res, GetUserID(users[j]))
			}
			_ = res
		}
	}
	fmt.Println("done in", time.Since(start))
}

var (
	once  sync.Once
	users []*User
)

func makeUsers(n int) []*User {
	once.Do(func() {
		users = make([]*User, 10_000_000)
		for i := 0; i < 10_000_000; i++ {
			users[i] = &User{ID: strconv.Itoa(i)}
		}
	})
	return users[:n]
}
