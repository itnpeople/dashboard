// main.go

package main

import (
	"fmt"
	"time"
	//"github.com/kore3lab/dashboard/backend/router"
)

func main() {

	//config.Setup()
	//router.CreateUrlMappings()
	//router.Router.Run(":3001")

	fn := test()

	fn(1)
	time.Sleep(1 * time.Second)
	fn(1)
	time.Sleep(1 * time.Second)
	fn(1)
	time.Sleep(1 * time.Second)

}

func test() func(int) {

	a := &struct {
		name  string
		count int
	}{name: "honester", count: 1}

	s := "@"

	return func(count int) {
		fmt.Println("count(before)=", a.count, s)
		a.count = a.count + count
		s = fmt.Sprintf("@%d", a.count)
		fmt.Println("count(after)=", a.count, s)
	}

	//go func() {
	//	for i := 0; i < 3; i++ {
	//		a.count++
	//		fmt.Println("count=", a.count)
	//		time.Sleep(1 * time.Second)
	//	}
	//}()
}
