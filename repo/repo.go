package repo

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/libgit2/git2go"
)

const repoPath = "./"

func uncommittedChanges(repo *git.Repository) (bool, error) {
	opts := &git.StatusOptions{}
	opts.Flags = git.StatusOptIncludeUntracked

	statusList, err := repo.StatusList(opts)
	if err != nil {
		return false, err
	}

	count, err := statusList.EntryCount()
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

// OpenMachine checkouts the branch holding the machines machine.yaml
func OpenMachine(machine string) error {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	ok, err := uncommittedChanges(repo)
	if err != nil {
		return err
	}

	if ok {
		return fmt.Errorf("Cannot open %s, there are uncommitted changes", machine)
	}

	// change HEAD ref to point to the machine's branch
	_, err = repo.References.CreateSymbolic("HEAD", "refs/heads/"+machine, true, "")
	if err != nil {
		return err
	}

	opts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing,
	}

	// checkout the head now pointing to the machine's branch
	if err := repo.CheckoutHead(opts); err != nil {
		return err
	}

	return nil
}

// CreateMachine creates a new git branch that represents a machine
func CreateMachine(machine string) error {
	refname := "refs/heads/" + machine

	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	ok, err := uncommittedChanges(repo)
	if err != nil {
		return err
	}

	if ok {
		return fmt.Errorf("Cannot create %s, there are uncommitted changes", machine)
	}

	// create a bare branch as outlined here: https://people.debian.org/~mika/forensics/git.html#branch
	_, err = repo.References.CreateSymbolic("HEAD", refname, true, "")
	if err != nil {
		return err
	}

	os.Remove(".git/index")
	files, err := filepath.Glob("*")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file == ".git" {
			continue
		}

		os.RemoveAll(file)
	}

	f, err := os.Create(machine + ".yaml")
	if err != nil {
		return err
	}
	f.Close()

	idx, err := repo.Index()
	if err != nil {
		return err
	}

	idx.AddByPath(machine + ".yaml")

	oid, err := idx.WriteTree()
	if err != nil {
		return err
	}

	err = idx.Write()
	if err != nil {
		return err
	}

	tree, err := repo.LookupTree(oid)
	if err != nil {
		return err
	}

	signature := &git.Signature{
		Name:  "Alterant",
		Email: "https://github.com/autonomy/alterant",
		When:  time.Now(),
	}

	message := "Add machine: " + machine

	// leave out the `parents` since this is an orphaned branch
	_, err = repo.CreateCommit(refname, signature, signature, message, tree)
	if err != nil {
		return err
	}

	return nil
}

// ListMachines lists the available machines in a repo
func ListMachines() error {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return err
	}

	iter, err := repo.NewBranchIterator(git.BranchLocal)
	if err != nil {
		return err
	}

	listMachines := func(b *git.Branch, t git.BranchType) error {
		name, err := b.Name()
		if err != nil {
			return err
		}

		fmt.Println(name)

		return nil
	}

	iter.ForEach(listMachines)

	return nil
}

// CurrentMachine returns the current branch
func CurrentMachine() (machine string, err error) {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	branch := head.Branch()
	if err != nil {
		return "", err
	}

	branchName, err := branch.Name()
	if err != nil {
		return "", err
	}

	return branchName, nil
}

// CloneToAlterantDir clones the requested machine to ~/.alterant
func CloneToAlterantDir(url string, machine string, alterantDir string) error {
	repoPath := path.Join(alterantDir, machine)
	_, err := git.Clone(url, repoPath, &git.CloneOptions{CheckoutBranch: machine})
	if err != nil {
		return err
	}

	return nil
}
