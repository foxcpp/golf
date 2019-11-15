package main

import (
	"debug/elf"
	"fmt"
	"log"
	"os"
)

type PkgInfo struct {
	Funcs uint64
	Objs  uint64
}

type SizeInfo struct {
	ActualSize uint64

	TotalFuncs   uint64
	TotalObjs    uint64
	UnknownFuncs uint64
	UnknownObjs  uint64

	RelocationData uint64
	DebugInfo      uint64
	SymbolTables   uint64
	GoLineTab      uint64

	Misc uint64

	Pkgs map[string]*PkgInfo
}

var debugSections = []string{
	".zdebug_aranges",
	".zdebug_pubnames",
	".zdebug_info",
	".zdebug_abbrev",
	".zdebug_line",
	".zdebug_frame",
	".zdebug_str",
	".zdebug_loc",
	".zdebug_pubtypes",
	".zdebug_ranges",
}

func analyze(f *os.File) (SizeInfo, error) {
	info := SizeInfo{
		Pkgs: make(map[string]*PkgInfo),
	}

	finfo, err := os.Stat(f.Name())
	if err != nil {
		return SizeInfo{}, fmt.Errorf("analyze %s: %w", f.Name(), err)
	}
	info.ActualSize = uint64(finfo.Size())

	noBitsSect := make(map[elf.SectionIndex]struct{})

	elfF, err := elf.NewFile(f)
	if err != nil {
		return SizeInfo{}, fmt.Errorf("analyze %s: %w", f.Name(), err)
	}

	for i, sect := range elfF.Sections {
		switch sect.Name {
		case ".gopclntab":
			info.GoLineTab = sect.Size
			continue
		case ".plt", ".plt.got", ".got", ".got.plt":
			info.RelocationData += sect.FileSize
			if *verbose {
				log.Println("counting", sect.Name, "as relocation data")
			}
			continue
		}
		switch sect.Type {
		case elf.SHT_RELA:
			info.RelocationData += sect.FileSize
			if *verbose {
				log.Println("counting", sect.Name, "as relocation data")
			}
		case elf.SHT_NOBITS:
			noBitsSect[elf.SectionIndex(i)] = struct{}{}
			if *verbose {
				log.Println("ignoring", sect.Name)
			}
		case elf.SHT_DYNSYM, elf.SHT_SYMTAB, elf.SHT_STRTAB:
			if *verbose {
				log.Println("counting", sect.Name, "as symbol table section")
			}
			info.SymbolTables += sect.FileSize
		}
	}

	for _, name := range debugSections {
		sect := elfF.Section(name)
		if sect == nil {
			continue
		}
		info.DebugInfo += sect.FileSize
	}

	syms, err := elfF.Symbols()
	if err != nil {
		if err == elf.ErrNoSymbols {
			log.Println("WARNING: No symbol table, per-package information is unavailable.")
		} else {
			return SizeInfo{}, fmt.Errorf("analyze %s: %w", f.Name(), err)
		}
	}

symloop:
	for _, sym := range syms {
		if sym.Size == 0 {
			continue
		}
		if _, ok := noBitsSect[sym.Section]; ok {
			continue
		}

		if *verbose {
			log.Printf("considering sym %v\n    size %v, sect %d, other %#x, info %#x",
				sym.Name, sym.Size, sym.Section, sym.Other, sym.Info)
		}

		sect := elfF.Sections[sym.Section]

		executable := (elfF.Sections[sym.Section].Flags & elf.SHF_EXECINSTR) == elf.SHF_EXECINSTR

		switch sect.Name {
		case ".gopclntab", ".plt", ".plt.got", ".got", ".got.plt":
			log.Printf("Warning: symbol %s is in section we already counted (%s), skipping", sym.Name, sect.Name)
			continue symloop
		}
		switch sect.Type {
		case elf.SHT_DYNSYM, elf.SHT_SYMTAB, elf.SHT_STRTAB, elf.SHT_RELA, elf.SHT_NOBITS:
			log.Printf("Warning: symbol %s is in section we already counted (%s), skipping", sym.Name, sect.Name)
			continue symloop
		}
		for _, debugSect := range debugSections {
			if sect.Name == debugSect {
				log.Printf("Warning: symbol %s is in section we already counted (%s), skipping", sym.Name, sect.Name)
				continue symloop
			}
		}

		if executable {
			info.TotalFuncs += sym.Size
		} else {
			info.TotalObjs += sym.Size
		}

		pkgName := guessPackage(sym.Name)

		if pkgName == "" {
			if *unknownList {
				log.Printf("Unknown symbol: %s (size: %d)", sym.Name, sym.Size)
			}
			if executable {
				info.UnknownFuncs += sym.Size
			} else {
				info.UnknownObjs += sym.Size
			}
			continue
		}

		if _, ok := info.Pkgs[pkgName]; !ok {
			info.Pkgs[pkgName] = &PkgInfo{}
		}

		if executable {
			info.Pkgs[pkgName].Funcs += sym.Size
		} else {
			info.Pkgs[pkgName].Objs += sym.Size
		}
	}

	info.Misc = info.ActualSize -
		info.TotalFuncs -
		info.TotalObjs -
		info.RelocationData -
		info.DebugInfo -
		info.SymbolTables -
		info.GoLineTab

	return info, nil
}
