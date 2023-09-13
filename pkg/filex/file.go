package filex

import (
	"bufio"
	"os"
)

// ReadLines 读取所有行
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0)
	fScanner := bufio.NewScanner(file)
	for fScanner.Scan() {
		lines = append(lines, fScanner.Text())
	}
	return lines, nil
}
