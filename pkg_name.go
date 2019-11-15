package main

import (
	"strings"
	"unicode"
)

var topStdlib = map[string]struct{}{
	"archive":   {},
	"bufio":     {},
	"builtin":   {},
	"bytes":     {},
	"cmd":       {},
	"compress":  {},
	"container": {},
	"context":   {},
	"crypto":    {},
	"database":  {},
	"debug":     {},
	"encoding":  {},
	"errors":    {},
	"expvar":    {},
	"flag":      {},
	"fmt":       {},
	"go":        {},
	"hash":      {},
	"html":      {},
	"image":     {},
	"index":     {},
	"internal":  {},
	"io":        {},
	"log":       {},
	"math":      {},
	"mime":      {},
	"net":       {},
	"os":        {},
	"path":      {},
	"plugin":    {},
	"reflect":   {},
	"regexp":    {},
	"runtime":   {},
	"sort":      {},
	"strconv":   {},
	"strings":   {},
	"sync":      {},
	"syscall":   {},
	"testdata":  {},
	"testing":   {},
	"text":      {},
	"time":      {},
	"unicode":   {},
	"unsafe":    {},
	"vendor":    {},
}

// based on debug/gosym.(*Sym).PackageName

func trimArraySize(name string) string {
	// Example: type..eq.[6][2]vendor/golang.org/x/text/secure/bidirule.ruleTransition
	// We are working with part:
	// [6][2]vendor/golang.org/x/text/secure/bidirule.ruleTransition

	if !strings.HasPrefix(name, "[") {
		return name
	}

	// This is loose and will also cut invalid permutations such as [[] 11112
	for i, ch := range name {
		switch {
		case unicode.IsDigit(ch), ch == '[', ch == ']':
		default:
			return name[i:]
		}
	}
	return name
}

func guessPackage(name string) string {
	if strings.HasPrefix(name, "go.itab.") {
		// Interface conversion table.
		// Example: go.itab.*flag.uint64Value,flag.Value
		name = strings.TrimPrefix(name, "go.itab.")
		name = strings.TrimPrefix(name, "*") // pointer...
		parts := strings.Split(name, ",")

		name = parts[0]
	} else if strings.HasPrefix(name, "type..") {
		// Equality / hash functions.
		name = strings.TrimPrefix(name, "type..")
		name = strings.TrimPrefix(name, "eq.") // or ...
		name = strings.TrimPrefix(name, "hash.")

		// Type definition is embedded into symbol name, there is no package name.
		if strings.HasPrefix(name, "struct") || strings.HasPrefix(name, "interface") {
			return ""
		}

		name = trimArraySize(name)

		if name == "string" {
			return "builtin"
		}
	}

	// A prefix of "type." and "go." is a compiler-generated symbol that doesn't belong to any package.
	if strings.HasPrefix(name, "go.") || strings.HasPrefix(name, "type.") {
		return "builtin"
	}

	if strings.HasPrefix(name, "main") {
		return "main"
	}

	pathend := strings.LastIndex(name, "/")
	if pathend < 0 {
		pathend = 0
	}

	if i := strings.Index(name[pathend:], "."); i != -1 {
		path := name[:pathend+i]
		slashes := strings.Count(path, "/")
		if _, std := topStdlib[path]; slashes == 0 && !std {
			// Filter out C functions.
			return ""
		}

		// Cut everything after receiver.
		pathend = strings.Index(path, ".(")
		if pathend < 0 {
			pathend = len(path)
		}

		return path[:pathend]
	}

	return ""
}
