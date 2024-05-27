package git

import (
	"fmt"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ProductionDeployments *git.Repository
	GitOps                *git.Repository
)

type Cloner struct {
	URL        string
	Repository string
	Branch     string
	CloneDir   string
}

// https://github.com/go-git/go-git/blob/master/_examples/branch/main.go
func Branch(repo *git.Repository, branchName string) (*plumbing.Reference, error) {
	headRef, err := repo.Head()
	if err != nil {
		return nil, err
	}
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	ref := plumbing.NewHashReference(plumbing.ReferenceName(branchRefName), headRef.Hash())
	err = repo.Storer.SetReference(ref)
	return nil, err
}

// https://github.com/go-git/go-git/blob/master/_examples/checkout-branch/main.go
func Checkout(repo *git.Repository, branchName string) error {
	_, err := repo.Head()
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	branchRefName := plumbing.NewBranchReferenceName(branchName)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(branchRefName),
		Force:  true,
	})
	if err != nil {
		return err
	}
	_, err = repo.Head()
	if err != nil {
		return err
	}
	return nil
}

// https://github.com/go-git/go-git/blob/master/_examples/clone/main.go
func Clone(c *Cloner) (*git.Repository, error) {
	// TODO: Research if it's feasible to clone this into memory and thus
	// avoid checking for its presence on disk.
	// https://github.com/go-git/go-git?tab=readme-ov-file#in-memory-example
	_, err := os.Stat(c.CloneDir)
	if !os.IsNotExist(err) {
		err = os.RemoveAll(c.CloneDir)
		if err != nil {
			return nil, err
		}
	}
	return git.PlainClone(c.CloneDir, false, &git.CloneOptions{
		URL:           fmt.Sprintf("%s/%s.git", c.URL, c.Repository),
		Progress:      nil,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", c.Branch)),
	})
}

// https://github.com/go-git/go-git/blob/master/_examples/commit/main.go
func Commit(repo *git.Repository) error {
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(".")
	if err != nil {
		return err
	}
	_, err = w.Status()
	if err != nil {
		return err
	}
	//	fmt.Println(status)
	_, err = w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "GitOps Bot",
			Email: "gitops-bot@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	//	obj, err := repo.CommitObject(commit)
	return nil
}

// https://github.com/go-git/go-git/blob/master/_examples/push/main.go
func Push(repo *git.Repository) error {
	return repo.Push(&git.PushOptions{})
}
