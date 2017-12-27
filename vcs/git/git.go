package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Repo struct {
	dir      string
	worktree string

	mu *sync.Mutex
}

func NewRepo(gitdir string) *Repo {
	return &Repo{
		dir:      gitdir,
		worktree: filepath.Dir(gitdir),
		mu:       new(sync.Mutex),
	}
}

func (git *Repo) Check(wg *sync.WaitGroup, fetch bool) {
	git.mu.Lock()
	defer git.mu.Unlock()
	defer wg.Done()

	gdarg := "--git-dir=" + git.dir
	wtarg := "--work-tree=" + git.worktree

	cmdGitBranch := exec.Command("git", wtarg, gdarg, "branch", "--no-color")
	cmdGitFetch := exec.Command("git", wtarg, gdarg, "fetch", "--all")
	cmdGitStatus := exec.Command("git", wtarg, gdarg, "status", "--porcelain")
	cmdGitRevlist := exec.Command("git", wtarg, gdarg, "rev-list", "--left-right", "--boundary", "@{upstream}...")

	if fetch {
		if fetchOut, err := cmdGitFetch.CombinedOutput(); err != nil {
			fmt.Println(git.worktree)
			os.Stdout.Write(fetchOut)
			fmt.Println("git fetch:", err)
			return
		}
	}

	statusOut, err := cmdGitStatus.CombinedOutput()
	if err != nil {
		fmt.Println(git.worktree)
		os.Stdout.Write(statusOut)
		fmt.Println("git status:", err)
		return
	}

	revlistOut, err := cmdGitRevlist.CombinedOutput()
	if err != nil {
		fmt.Println(git.worktree)
		os.Stdout.Write(revlistOut)
		fmt.Println("git rev-list:", err)
		return
	}

	branchOut, err := cmdGitBranch.CombinedOutput()
	if err != nil {
		fmt.Println(git.worktree)
		os.Stdout.Write(revlistOut)
		fmt.Println("git branch:", err)
		return
	}

	if len(statusOut) > 0 || len(revlistOut) > 0 {
		fmt.Fprintln(os.Stdout, strings.Repeat("-", 80))
		fmt.Fprintln(os.Stdout, git.worktree)

		fmt.Fprintln(os.Stdout) // print extra new line
		fmt.Fprintln(os.Stdout, "Branches:")
		os.Stdout.Write(branchOut)

		fmt.Fprintln(os.Stdout) // print extra new line
		fmt.Fprintln(os.Stdout, "Status:")
		if len(statusOut) > 0 {
			os.Stdout.Write(statusOut)
		} else {
			fmt.Fprintln(os.Stdout, "OK")
		}

		fmt.Fprintln(os.Stdout) // print extra new line
		fmt.Fprintln(os.Stdout, "Remote:")
		if len(revlistOut) > 0 {
			os.Stdout.Write(revlistOut)
		} else {
			fmt.Fprintln(os.Stdout, "OK")
		}
		fmt.Fprintln(os.Stdout, strings.Repeat("-", 80))
	}
	return
}
