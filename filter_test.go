package main

import (
    "fmt"
    "testing"
)

func TestFilter(t*testing.T) {
    a := filter([]string{"a","b","c"}, func(v string) bool { return v != "b" })
    if fmt.Sprint(a) != "[a c]" {
        t.Errorf("bad result: %v", a)
    }
}
