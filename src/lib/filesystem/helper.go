package filesystem

import (
	"os"
	"log"
	"time"
	"path/filepath"
	"io/ioutil"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/chrootlogin/go-wiki/src/repo"
)

func readFile(fs *filesystem, path string) (*File, error) {
	// Get FileInfo
	fileinfo, err := fs.Filesystem.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileinfo.IsDir() {
		return nil, ErrIsDir
	}

	// Open file
	file, err := fs.Filesystem.OpenFile(path, os.O_RDONLY, fs.FilePermissionMode)
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

	// Close file
	err = file.Close()
	if err != nil {
		log.Println("close file: " + err.Error())
		return nil, err
	}

	return &File{
		Content: string(data),
		FileInfo: fileinfo,
	}, nil
}

func commitFile(fs *filesystem, path string, data []byte, commit repo.Commit) error {
	// Open file
	file, err := fs.Filesystem.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fs.FilePermissionMode)
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
	if len(fs.Chroot) > 0 {
		_, err = fs.Worktree.Add(filepath.Join(fs.Chroot, path))
	} else {
		_, err = fs.Worktree.Add(path)
	}
	if err != nil {
		log.Println("adding file: " + err.Error())
		return err
	}

	// Creating commit
	_, err = fs.Worktree.Commit(commit.Message, &git.CommitOptions{
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