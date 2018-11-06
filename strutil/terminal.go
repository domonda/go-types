package strutil

import (
	"fmt"
	"strings"
)

func ConfirmQuestionOnTerminal(format string, v ...interface{}) bool {
	for {
		fmt.Printf(format+" [y/n]:", v...)
		var input string
		fmt.Scanln(&input)
		switch strings.ToUpper(input) {
		case "Y":
			return true
		case "N":
			return false
		}
		fmt.Println("Invalid choice, repeat answer:")
	}
}
