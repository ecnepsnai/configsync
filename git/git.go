package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Git describes a git instance
type Git struct {
	gitPath string
	repoDir string
}

const minimumGitVersion = 170

// New initalize a new git instance
func New(gitBinPath string, repoDir string) (*Git, error) {
	g := Git{
		gitPath: gitBinPath,
		repoDir: repoDir,
	}

	version, err := g.Version()
	if err != nil {
		return nil, err
	}
	if *version < minimumGitVersion {
		return nil, fmt.Errorf("unsupported git version (too old)")
	}

	return &g, nil
}

func (g *Git) exec(verb string, args ...string) ([]byte, error) {
	args = append([]string{verb}, args...)
	cmd := exec.Command(g.gitPath, args...)
	cmd.Dir = g.repoDir
	return cmd.CombinedOutput()
}

// Version the git binary version
func (g *Git) Version() (*int, error) {
	out, err := g.exec("version")
	if err != nil {
		return nil, err
	}
	pattern := regexp.MustCompile("\\d+\\.\\d+\\.\\d+(\\.\\d+)?")
	versionString := pattern.FindString(string(out))
	if versionString == "" {
		return nil, fmt.Errorf("unknown git version: %s", out)
	}

	versionString = strings.ReplaceAll(versionString, ".", "")
	version, err := strconv.Atoi(versionString)
	if err != nil {
		return nil, fmt.Errorf("unknown git version: %s", versionString)
	}

	return &version, nil
}

// InitIfNeeded initalize a new repo if needed
func (g *Git) InitIfNeeded() error {
	_, err := g.exec("status")
	if err == nil {
		return nil
	}
	_, err = g.exec("init")
	if err != nil {
		return err
	}

	return nil
}

// CurrentBranch get the current branch
func (g *Git) CurrentBranch() (*string, error) {
	out, err := g.exec("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, err
	}
	branch := string(out)
	branch = strings.ReplaceAll(branch, "\n", "")
	return &branch, nil
}

// Checkout checkout a specific branch
func (g *Git) Checkout(branch string) error {
	current, err := g.CurrentBranch()
	if err != nil {
		return err
	}
	if branch == *current {
		return nil
	}

	_, err = g.exec("checkout", "-b", branch)
	if err != nil {
		return err
	}
	return nil
}

// Pull perform a git pull
func (g *Git) Pull() error {
	_, err := g.exec("pull")
	if err != nil {
		return err
	}
	return nil
}

// Push perform a git push
func (g *Git) Push(remote, local string) error {
	_, err := g.exec("push", remote, local)
	if err != nil {
		return err
	}
	return nil
}

// Remove perform a git rm -f
func (g *Git) Remove(filePath string) error {
	_, err := g.exec("rm", "-f", filePath)
	if err != nil {
		return err
	}
	return nil
}

// HasChanges does the repo have any unstaged or untracked files
func (g *Git) HasChanges() bool {
	out, err := g.exec("status", "--porcelain")
	if err != nil {
		panic(err)
	}
	return len(out) > 1
}

// Add perform a git add
func (g *Git) Add(files ...string) error {
	_, err := g.exec("add", files...)
	if err != nil {
		return err
	}
	return nil
}

// Commit perform a git commit
func (g *Git) Commit(message string, author string) error {
	_, err := g.exec("commit", "-m", message, "--author", author)
	if err != nil {
		return err
	}
	return nil
}
