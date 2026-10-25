package interactive

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

func readString(scanner *bufio.Scanner, out io.Writer, prompt string) (string, error) {
	fmt.Fprint(out, prompt)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func readInt(scanner *bufio.Scanner, out io.Writer, prompt string, min, max int) (int, error) {
	for {
		line, err := readString(scanner, out, prompt)
		if err != nil {
			return 0, err
		}
		value, parseErr := strconv.Atoi(line)
		if parseErr != nil || value < min || value > max {
			fmt.Fprintf(out, "Ошибка: введите целое число от %d до %d.\n", min, max)
			continue
		}
		return value, nil
	}
}

func readFloat(scanner *bufio.Scanner, out io.Writer, prompt string) (float64, error) {
	for {
		line, err := readString(scanner, out, prompt)
		if err != nil {
			return 0, err
		}
		value, parseErr := strconv.ParseFloat(line, 64)
		if parseErr != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			fmt.Fprintln(out, "Ошибка: введите неотрицательное число.")
			continue
		}
		return value, nil
	}
}
