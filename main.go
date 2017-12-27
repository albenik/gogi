package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/albenik/gogi/vcs/git"
)

func mainf_() int {
	helpFlag := flag.Bool("help", false, "Display help")
	fetchFlag := flag.Bool("fetch", false, "Fetch from remotes and check commits ahead")
	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return 0
	}

	var (
		err  error
		root string
		wg   sync.WaitGroup
	)

	if flag.NArg() > 0 {
		root = flag.Arg(0)
	} else {
		if root, err = os.Getwd(); err != nil {
			fmt.Println(err)
			return 1
		}
	}
	if root, err = filepath.Abs(root); err != nil {
		fmt.Println(err)
		return 1
	}
	if _, err := os.Stat(root); err != nil {
		fmt.Println(err)
		return 1
	}

	err = filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("ERROR:", err)
			return err
		}
		if f.IsDir() && f.Name() == ".git" {
			wg.Add(1)
			go git.NewRepo(path).Check(&wg, *fetchFlag)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return 1
	}
	wg.Wait()
	return 0
}

func main() {
	os.Exit(mainf_())
}
