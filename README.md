golf
======

_Go & Obnoxious Lots of Fat_

Small utility to extract information about sections and symbols size from Go
ELF binaries.

Usage
-------

```
golf executable-path
```

Report example
----------------

```
Summary:
  Actual executable size (according to FS): 23.85 MiB
  Executable symbols size (total/unknown/package): 6.43 MiB / 877.06 KiB / 5.58 MiB
  Non-executable symbols size (total/unknown/package): 4.83 MiB / 72.84 KiB / 4.76 MiB
  Relocations (PIE): 3.31 MiB
  Debug info: 4.42 MiB
  Symbol table(s): 1.34 MiB
  Go source location info: 0.00 B

  Unknown: 3.52 MiB

Per-package statistics (total, executable, non-executable):
  runtime: 4.11 MiB / 360.37 KiB / 3.76 MiB
  net/http: 478.49 KiB / 459.16 KiB / 19.33 KiB
  github.com/miekg/dns: 424.33 KiB / 414.42 KiB / 9.90 KiB
  crypto/tls: 302.34 KiB / 290.07 KiB / 12.27 KiB
  net: 263.14 KiB / 253.06 KiB / 10.08 KiB
  golang.org/x/text/encoding/traditionalchinese: 229.04 KiB / 3.39 KiB / 225.65 KiB
  github.com/foxcpp/go-imap-sql: 186.78 KiB / 182.32 KiB / 4.46 KiB
  math/big: 164.70 KiB / 161.95 KiB / 2.75 KiB
  github.com/lib/pq: 121.92 KiB / 111.22 KiB / 10.70 KiB
  github.com/go-sql-driver/mysql: 119.45 KiB / 114.44 KiB / 5.01 KiB
```

golf vs goweight
------------------

goweight measures size of separate object files produced by Go toolchain. golf
analyzes structure of the final executable.

### golf
- Takes linker transformations into account
- Counts C functions as 'unknown' without per-package separation (limitation)
- Debug, file/line and relocation info is not reported in per-package sizes (can be fixed)
- Limited to ELF binary format (can be fixed)

## goweight
- Uses pre-linker artifacts, ignoring what linker can possibly do with them
- Associates functions from C with Go packages that introduced them
- Debug and file/line info is included intp per-package sizes
- Works with any binary format (e.g. Win32 PE)
- Requires source code 
