package utils

import (
	"bufio"
	"os"
	"strings"
)

func LoadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			// this is complete local purpose. on prods, its not allowed setting env this way but prooper export
			os.Setenv(parts[0], parts[1])
		}
	}
}
