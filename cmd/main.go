package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"github.com/JoaoBraveCoding/prom-storage-analysis/pkg/parser"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	file, err := os.Open(os.Args[1])
	check(err)
	defer file.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var line string
	var expression string
	var expressions []string
	for {
		line, err = reader.ReadString('\n')
		if len(line) != 1 {
			expression += line
		} else {
			if expression != "" {
				expressions = append(expressions, expression)
				expression = ""
			}
		}

		if err != nil {
			expressions = append(expressions, expression)
			break
		}
	}

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}

	mapOfMetrics := make(map[string]bool)
	for _, expr := range expressions {
		metricsInExpr := parser.GetMetrics(expr)
		for _, metric := range metricsInExpr {
			if _, ok := mapOfMetrics[metric]; !ok {
				mapOfMetrics[metric] = true
			}
		}
	}

	for metric := range mapOfMetrics{
		fmt.Println(metric)
	}
}
