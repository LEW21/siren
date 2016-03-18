package main

import "os"

func EnsureSirenDirExists() {
	_, err := os.Stat("/var/lib/siren")
	if err != nil {
		err := os.Mkdir("/var/lib/siren", 0700)
		if err != nil {
			panic(err)
		}
	}
}
