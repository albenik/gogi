package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
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
			go func(p string) {
				defer wg.Done()
				if err := check(p, *fetchFlag); err != nil {
					fmt.Println()
					fmt.Println("ERROR:", err)
				}
			}(filepath.Join(path, ".."))
			// go g.NewRepo(path).Check(&wg, *fetchFlag)
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

func check(path string, fetch bool) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("cannot open repo %q: %w", path, err)
	}

	if fetch {
		remotes, err := repo.Remotes()
		if err != nil {
			return fmt.Errorf("cannot get remotes %q: %w", path, err)
		}

		for _, r := range remotes {
			err = r.Fetch(&git.FetchOptions{})
			if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
				return fmt.Errorf("cannot fetch remote %q → %q: %w", path, r, err)
			}
		}
	}

	tree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("cannot get worktree %q: %w", path, err)
	}

	status, err := tree.Status()
	if err != nil {
		return fmt.Errorf("cannot get status %q: %w", path, err)
	}

	ign, err := gitignore.LoadGlobalPatterns(osfs.New("/"))
	if err != nil {
		return fmt.Errorf("cannot read global gitinore file: %w", err)
	}

	fmt.Println("Status:")
	for fn, fs := range status {
		stat, err := os.Stat(fn)
		if err != nil {
			return fmt.Errorf("stat error: %q: %w", fn, err)
		}
		for _, i := range ign {
			fmt.Println(fn, i, i.Match([]string{fn}, stat.IsDir()))
			if i.Match([]string{fn}, stat.IsDir()) == gitignore.Exclude {
				continue
			}
		}
		fmt.Println(string(fs.Worktree), string(fs.Staging), fn)
	}

	// head, err := repo.Head()
	// if err != nil {
	// 	return fmt.Errorf("cannot get head %q: %w", path, err)
	// }
	// head.Name()
	return nil
}
