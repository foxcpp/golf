package main

import "strconv"

func humanSize(b uint64) string {
	suffix := " B"
	bf := float64(b)

	if bf > 1024 {
		bf /= 1024
		suffix = " KiB"
	}
	if bf > 1024 {
		bf /= 1024
		suffix = " MiB"
	}
	if bf > 1024 {
		bf /= 1024
		suffix = " GiB"
	}

	return strconv.FormatFloat(bf, 'f', 2, 64) + suffix
}
