package parser

import (
	"bufio"
	"os"
)

// ReadLines given a file path it will read each line of the file to an entry in
// a slice of strings where in the first entry will be the first line of the
// file and on the last the last line of the file
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
