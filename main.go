package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
)

var (
	verbose     = flag.Bool("v", false, "Log what's going on")
	pkgsLimit   = flag.Int("top", 10, "Print stats only for first N packages, 0 to disable")
	jsonOut     = flag.Bool("json", false, "Print all stats in JSON format")
	unknownList = flag.Bool("unknown-list", false, "Print unclassified (unknown) symbol names")
)

func main() {
	flag.Parse()

	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	if len(flag.Args()) != 1 {
		log.Fatalln("Usage:", os.Args[0], "[options] <file>")
	}

	file, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	sizeInfo, err := analyze(file)
	if err != nil {
		log.Fatalln(err)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", " ")
		if err := enc.Encode(sizeInfo); err != nil {
			log.Fatalln(err)
		}
		return
	}

	fmt.Println("Summary:")
	fmt.Println("  Actual executable size (according to FS):", humanSize(sizeInfo.ActualSize))
	fmt.Printf("  Executable symbols size (total/unknown/package): %s / %s / %s\n",
		humanSize(sizeInfo.TotalFuncs), humanSize(sizeInfo.UnknownFuncs), humanSize(sizeInfo.TotalFuncs-sizeInfo.UnknownFuncs))
	fmt.Printf("  Non-executable symbols size (total/unknown/package): %s / %s / %s\n",
		humanSize(sizeInfo.TotalObjs), humanSize(sizeInfo.UnknownObjs), humanSize(sizeInfo.TotalObjs-sizeInfo.UnknownObjs))
	fmt.Println("  Relocations:", humanSize(sizeInfo.RelocationData))
	fmt.Println("  Debug info:", humanSize(sizeInfo.DebugInfo))
	fmt.Println("  Symbol table(s):", humanSize(sizeInfo.SymbolTables))
	fmt.Println("  Go source location info:", humanSize(sizeInfo.GoLineTab))
	fmt.Println()
	fmt.Println("  Unknown:", humanSize(sizeInfo.Misc))
	fmt.Println()
	fmt.Println("Per-package statistics (total, executable, non-executable):")

	sortedPkgs := make([]string, 0, len(sizeInfo.Pkgs))
	for k := range sizeInfo.Pkgs {
		sortedPkgs = append(sortedPkgs, k)
	}
	sort.Slice(sortedPkgs, func(i, j int) bool {
		iInfo := sizeInfo.Pkgs[sortedPkgs[i]]
		jInfo := sizeInfo.Pkgs[sortedPkgs[j]]
		// Sort in reverse.
		return (iInfo.Funcs + iInfo.Objs) > (jInfo.Funcs + jInfo.Objs)
	})

	for i, pkg := range sortedPkgs {
		info := sizeInfo.Pkgs[pkg]

		if *pkgsLimit != 0 && i == *pkgsLimit {
			break
		}

		fmt.Printf("  %s: %s / %s / %s\n", pkg, humanSize(info.Funcs+info.Objs), humanSize(info.Funcs), humanSize(info.Objs))
	}
}
