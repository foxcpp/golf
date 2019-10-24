package main

import (
	"debug/elf"
	"strings"
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

func guessPackage(sym elf.Symbol) string {
	name := sym.Name

	// A prefix of "type." and "go." is a compiler-generated symbol that doesn't belong to any package.

	if strings.HasPrefix(name, "go.") || strings.HasPrefix(name, "type.") {
		return ""
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
