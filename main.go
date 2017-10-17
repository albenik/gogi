package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func checkRepo(path string, fetchFlag bool, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	dir := filepath.Dir(path)
	worktree := "--work-tree=" + dir
	gitdir := "--git-dir=" + path

	fetchcmd := exec.Command("git", worktree, gitdir, "fetch", "--all")
	statuscmd := exec.Command("git", worktree, gitdir, "status", "--porcelain")
	revlistcmd := exec.Command("git", worktree, gitdir, "rev-list", "HEAD@{upstream}..HEAD")

	var statusOut, revlistOut []byte
	var err error

	if fetchFlag {
		if fetchOut, err := fetchcmd.CombinedOutput(); err != nil {
			mu.Lock()
			defer mu.Unlock()
			fmt.Println(dir)
			os.Stdout.Write(fetchOut)
			fmt.Println("git fetch:", err)
			return
		}
	}
	if statusOut, err = statuscmd.CombinedOutput(); err != nil {
		mu.Lock()
		defer mu.Unlock()
		fmt.Println(dir)
		os.Stdout.Write(statusOut)
		fmt.Println("git status:", err)
		return
	}
	if len(statusOut) == 0 {
		if revlistOut, err = revlistcmd.CombinedOutput(); err != nil {
			mu.Lock()
			defer mu.Unlock()
			fmt.Println(dir)
			os.Stdout.Write(revlistOut)
			fmt.Println("git rev-list:", err)
			return
		}
	}

	mu.Lock()
	defer mu.Unlock()

	if len(statusOut) > 0 || len(revlistOut) > 0 {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, dir)
		os.Stdout.Write(statusOut)
		os.Stdout.Write(revlistOut)
	}
	return
}

func mainfunc() int {
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
		mu   sync.Mutex
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
		if f.IsDir() && f.Name() == ".git" {
			wg.Add(1)
			go checkRepo(path, *fetchFlag, &mu, &wg)
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
	os.Exit(mainfunc())
}
