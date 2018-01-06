package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	separator      = "───"
	dir_separator  = "├───"
	last_element   = "└───"
	dir_shift      = "│\t"
	last_dir_shift = "\t"
)

type Context struct {
	isFirst, isLast bool
	prefix          string
}

func prettyPrint(prefix string, out io.Writer, path string, info os.FileInfo, printFiles bool) {
	// Printing
	if !info.IsDir() && !printFiles {
		return
	}

	fmt.Fprintf(out, prefix)

	if !info.IsDir() {
		if printFiles {
			fmt.Fprintf(out, filepath.Base(path))
			if info.Size() > 0 {
				fmt.Fprintf(out, " (%vb)\n", info.Size())
			} else {
				fmt.Fprintf(out, " (empty)\n")
			}
		}
	} else {
		// Do not print initial dir
		if path != "." {
			fmt.Fprintf(out, filepath.Base(path))
			fmt.Fprintf(out, "\n")
		}
	}
}

func walkTree(walkPrefix string, out io.Writer, path string, printFiles bool) error {
	pattern := filepath.Join(path, "*")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	var dirs []string

	for _, d := range paths {
		stat, err := os.Stat(d)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}

		if printFiles || stat.IsDir() {
			dirs = append(dirs, d)
		}
	}

	var prefix string
	var dirPrefix string

	for i, d := range dirs {
		stat, err := os.Stat(d)
		if err != nil {
			return err
		}

		isLast := i == len(dirs)-1

		if !isLast {
			prefix = dir_separator
			dirPrefix = dir_shift
		} else {
			prefix = last_element
			dirPrefix = last_dir_shift
		}

		prettyPrint(walkPrefix+prefix, out, d, stat, printFiles)

		if stat.IsDir() {
			err = walkTree(walkPrefix+dirPrefix, out, d, printFiles)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	currentDir, err := os.Open(path)
	if err != nil {
		fmt.Println("Error ocurs in open", err)
	}

	stat, err := currentDir.Stat()
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return nil
	}

	walkTree("", out, path, printFiles)

	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
