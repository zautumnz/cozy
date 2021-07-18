package evaluator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var searchPaths []string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting cwd: %s", err)
	}

	if e := os.Getenv("COZYPATH"); e != "" {
		tokens := strings.Split(e, ":")
		for _, token := range tokens {
			addPath(token) // ignore errors
		}
	} else {
		searchPaths = append(searchPaths, cwd)
	}
}

func addPath(path string) error {
	path = os.ExpandEnv(filepath.Clean(path))
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	searchPaths = append(searchPaths, absPath)
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FindModule finds a module based on name, used by the evaluator
func FindModule(name string) string {
	basename := fmt.Sprintf("%s.cz", name)
	for _, p := range searchPaths {
		filename := filepath.Join(p, basename)
		if exists(filename) {
			return filename
		}
	}
	return ""
}

// Check to see if a value is a number
func IsNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)

	return err == nil
}
