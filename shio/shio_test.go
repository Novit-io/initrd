package shio

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

func ExampleOutput() {
	done := false
	defer func() { done = true }()
	go func() {
		time.Sleep(time.Second)
		if !done {
			panic("timeout")
		}
	}()

	shio := NewWithCap(3)

	r1 := shio.NewReader()

	// read as you write
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, r1)
		fmt.Println("-- r1 done --")
	}()

	fmt.Fprintln(shio, "hello1")
	fmt.Fprintln(shio, "hello2")
	shio.Close()

	wg.Wait()

	// read after close
	r2 := shio.NewReader()
	io.Copy(os.Stdout, r2)
	fmt.Println("-- r2 done --")

	// Output:
	// hello1
	// hello2
	// -- r1 done --
	// hello1
	// hello2
	// -- r2 done --
}
