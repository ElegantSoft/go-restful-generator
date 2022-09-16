package main

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

func Execute() {

	prompt := promptui.Prompt{
		Label: "package name",
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)
}

func main() {
	Execute()
}
