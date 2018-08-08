package repo

import (
	"log"
	"os"
	"time"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/chrootlogin/go-wiki/src/common"
)

func HasRaw(path string) bool {
	wt, _ := repo.Worktree()
	if _, err := wt.Filesystem.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func HasFile(path string) bool {
	path = filepath.Join("pages", path)

	return HasRaw(path)
}

func HasWithChroot(basedir string, path string) bool {
	wt, _ := repo.Worktree()
	fs, _ := wt.Filesystem.Chroot(basedir)

	if _, err := fs.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func GetRaw(path string) ([]byte, error) {
	// open workspace
	wt, err := repo.Worktree()
	if err != nil {
		log.Println("opening worktree: " + err.Error())
		return nil, err
	}

	// Open file
	file, err := wt.Filesystem.OpenFile(path, os.O_RDONLY, R_PERMS)
	if err != nil {
		log.Println("open file: " + err.Error())
		return nil, err
	}

	// Read file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("read file: " + err.Error())
		return nil, err
	}

	err = file.Close()
	if err != nil {
		log.Println("close file: " + err.Error())
		return nil, err
	}

	return data, nil
}

func GetFile(path string) (common.File, error) {
	path = filepath.Join("pages", path)

	/*commits := getCommitHistoryOfFile(path)
	fmt.Println(commits)*/

	data, err := GetRaw(path)
	if err != nil {
		return common.File{}, err
	}

	// Convert json to object
	var file = common.File{}
	err = json.Unmarshal(data, &file)
	if err != nil {
		log.Println("unmarshal: " + err.Error())
		return common.File{}, err
	}

	return file, nil
}

func GetWithChroot(basedir string, path string) ([]byte, error){
	// open workspace
	wt, err := repo.Worktree()
	if err != nil {
		log.Println("opening worktree: " + err.Error())
		return nil, err
	}

	// open chroot
	fs, err := wt.Filesystem.Chroot(basedir)
	if err != nil {
		log.Println("open chroot: " + err.Error())
		return nil, err
	}

	// Open file
	file, err := fs.OpenFile(path, os.O_RDONLY, R_PERMS)
	if err != nil {
		log.Println("open file: " + err.Error())
		return nil, err
	}

	// Read file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("read file: " + err.Error())
		return nil, err
	}

	err = file.Close()
	if err != nil {
		log.Println("close file: " + err.Error())
		return nil, err
	}

	return data, nil
}

func SaveWithChroot(basedir string, path string, data []byte, commit Commit) error {
	// open workspace
	wt, err := repo.Worktree()
	if err != nil {
		log.Println("opening worktree: " + err.Error())
		return err
	}

	// open chroot
	fs, err := wt.Filesystem.Chroot(basedir)
	if err != nil {
		log.Println("open chroot: " + err.Error())
		return err
	}

	// Open file
	file, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, R_PERMS)
	if err != nil {
		log.Println("open file: " + err.Error())
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		file.Close()
		log.Println("write to file: " + err.Error())
		return err
	}

	// close file
	err = file.Close()
	if err != nil {
		log.Println("close file: " + err.Error())
		return err
	}

	// Add file
	_, err = wt.Add(filepath.Join(basedir, path))
	if err != nil {
		log.Println("adding file: " + err.Error())
		return err
	}

	// Creating commit
	_, err = wt.Commit(commit.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  commit.Author.Username,
			Email: commit.Author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Println("commit: " + err.Error())
		return err
	}

	return nil
}

func SaveRaw(path string, data []byte, commit Commit) error {
	// open workspace
	wt, err := repo.Worktree()
	if err != nil {
		log.Println("opening worktree: " + err.Error())
		return err
	}

	// Open file
	file, err := wt.Filesystem.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, R_PERMS)
	if err != nil {
		log.Println("open file: " + err.Error())
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		file.Close()
		log.Println("write to file: " + err.Error())
		return err
	}

	// close file
	err = file.Close()
	if err != nil {
		log.Println("close file: " + err.Error())
		return err
	}

	// Add file
	_, err = wt.Add(path)
	if err != nil {
		log.Println("adding file: " + err.Error())
		return err
	}

	// Creating commit
	_, err = wt.Commit(commit.Message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  commit.Author.Username,
			Email: commit.Author.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Println("commit: " + err.Error())
		return err
	}

	return nil
}

func SaveFile(path string, file common.File, commit Commit) error {
	path = filepath.Join("pages", path)

	jsonBytes, err := json.Marshal(&file)
	if err != nil {
		log.Println("marshal: " + err.Error())
		return err
	}

	return SaveRaw(path, jsonBytes, commit)
}

func MkdirPage(path string) error {
	path = filepath.Join(repositoryPath, "pages", path)

	return os.MkdirAll(path, os.ModePerm);
}

func getCommitHistoryOfFile(path string) []*object.Commit {
	objects := []*object.Commit{}

	ref, err := repo.Head()
	if err == nil {
		cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err == nil {
			cIter.ForEach(filterByChangesToPath(repo, path, func(c *object.Commit) error {
				objects = append(objects, c)
				return nil
			}))
		}
	}

	return objects
}