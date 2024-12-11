package main

import (
	"fmt"
	"os"
)

func ExampleParseEnv_setVariable() {
	os.Setenv("env", "Given variable")
	fmt.Println(ParseEnv("env", "fallback"))
	// Output: Given variable
}

func ExampleParseEnv_unsetVariable() {
	os.Unsetenv("env")
	fmt.Println(ParseEnv("env", "Fallback"))
	// Output: Fallback
}
