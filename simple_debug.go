package main

import (
"fmt"
"os"
)

func main() {
fmt.Println("Current working directory:")
pwd, _ := os.Getwd()
fmt.Println(pwd)
}
