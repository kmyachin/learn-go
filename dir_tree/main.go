package main

import (
	"os"
	"fmt"
	"io"
	"path/filepath"
)

const separator  = "───"
const dir_separator = "├───"
const last_element = "└───"
const dir_shift = "│\t"
const last_dir_shift = "\t"

type Context struct {
	isFirst, isLast bool
	prefix string
}



func prettyPrint(context Context, out io.Writer, path string, info os.FileInfo, printFiles bool) {
	// Printing
	if !info.IsDir() && !printFiles {
		return
	}

	fmt.Fprintf(out, context.prefix)

	var prefix string
	if !context.isLast{
		prefix = dir_separator
	} else {
		prefix = last_element
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

func walkTree(context Context, out io.Writer, path string, printFiles bool) error {
	pattern := filepath.Join(path,"*")
	paths,err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	prefix := context.prefix
	if !context.isFirst {
		if context.isLast {
			context.prefix += last_dir_shift
		} else {
			context.prefix += dir_shift
		}
	}

	//is_first := context.isFirst
	var dirs []string

	for _, d := range paths {
		stat, err := os.Stat(d)
		if err != nil {
			fmt.Fprintln(out, err)
			continue
		}

		if printFiles || stat.IsDir(){
			dirs = append(dirs, d)
		}
	}

	is_last := context.isLast
	for i, d := range dirs {
		context.isLast = false

		stat, err := os.Stat(d)
		if err != nil {
			return err
		}


		if i == len(dirs)-1 {
			context.isLast = true
		}
		is_first := context.isFirst

		context.isFirst = false
		prettyPrint(context, out, d, stat, printFiles)
		if (stat.IsDir()) {
			err = walkTree(context, out, d, printFiles)
		}
		context.isFirst = is_first

		if err != nil {
			return err
		}
	}
	context.isLast = is_last
	context.prefix = prefix

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	var context Context
	context.isFirst = true
	currentDir,err := os.Open(path)
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

	walkTree(context, out, path, printFiles)

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
