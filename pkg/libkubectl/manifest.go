package libkubectl

import "strings"

func manifestFilesToArgs(manifestFiles []string) []string {
	args := []string{}
	for _, path := range manifestFiles {
		args = append(args, "-f", strings.TrimSpace(path))
	}
	return args
}
