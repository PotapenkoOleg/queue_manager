package version

import (
	"fmt"
	"strings"
)

func PrintSeparator() {
	fmt.Println(strings.Repeat("*", 80))
}

func PrintBanner() {
	fmt.Println(
		fmt.Sprintf(
			"%s version %d.%d.%d (%s)",
			ProductName, VersionMajor, VersionMinor, VersionPatch, VersionAlias,
		),
	)
	fmt.Printf("License: %s\n", License)
	fmt.Printf("Link: %s\n", Link)
	fmt.Printf("Copyright © %s. %s\n", Copyright, CopyrightYears)
}
