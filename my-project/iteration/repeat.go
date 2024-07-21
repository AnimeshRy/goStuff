package iteration

import "fmt"

func Repeat(character string, iteration_count int) string {
	var repeated string
	for i := 0; i < iteration_count; i++ {
		repeated = repeated + character
	}
	return repeated
}

func main() {
	fmt.Println("sdssad")
}

