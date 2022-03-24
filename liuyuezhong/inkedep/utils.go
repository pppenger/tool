package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/tools/go/vcs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	DotGit = ".git"
)

// FindGitRoot searches from the starting directory upwards looking for a
// .git directory until we get to the root of the filesystem.
func FindGitRoot(from string) (string, error) {
	t := from
	for {
		mp := filepath.Join(from, DotGit)
		info, err := os.Stat(mp)
		if err == nil && info.IsDir() {
			return from, nil
		} else if err == nil && !info.IsDir() {
			return "", fmt.Errorf("%s is a file, not a valid %s directory", DotGit, from)
		}
		if !os.IsNotExist(err) {
			// Some err other than non-existence - return that out
			return "", err
		}

		parent := filepath.Dir(from)
		if parent == from {
			return "", fmt.Errorf("could not find %s directory from %s", DotGit, t)
		}
		from = parent
	}
}

func FindRepoRoot(path string) (string, string, error) {
	repo, err := vcs.RepoRootForImportPath(path, false)
	if err == nil {
		return repo.Repo, repo.Root, nil
	}
	return "", "", err
}

func FindRepoRootDynamic(path string) (string, error) {
	repo, err := vcs.RepoRootForImportDynamic(path, false)
	if err == nil {
		return repo.Repo, nil
	}
	return "", err
}

func FindRepoCacheRoot(path string) (string, string, error) {
	path = strings.Replace(path, ".git", "", 1)
	repo := fmt.Sprintf("git.inke.cn/cache/%s", path)
	return FindRepoRoot(repo)
}

func IsPathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Interactive(fmts string, sep ...interface{}) bool {
	reader := bufio.NewReader(os.Stdin)
	doPrint(fmt.Sprintf("%s[Y/y] choice to true:", fmts), sep...)
	data, _, _ := reader.ReadLine()
	if string(data) == "Y" || string(data) == "y" {
		return true
	}
	return false
}

func Exec(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	return c.Run()
}

func Remove(dir string) error {
	if ok, _ := IsPathExist(dir); ok {
		doWarning("removing %s\n", dir)
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("remove err: %v\n", err)
			return err
		}
	}

	return nil
}

func Move(from, to string) {
	if ok, _ := IsPathExist(from); ok {
		doWarning("moving %s to %s\n", from, to)
		if err := os.Rename(from, to); err != nil {
			fmt.Printf("move err: %v\n", err)
		}
	}
}

func baseLib(mod string, base map[string]bool) bool {
	for v := range base {
		if strings.Contains(mod, v) {
			return true
		}
	}
	return false
}

func exchagneModPath(modPath string) (string, bool) {
	if strings.HasPrefix(modPath, "github.com") ||
		strings.HasPrefix(modPath, "golang.org") {
		ss := strings.Split(modPath, "/")
		if len(ss) >= 3 {
			return strings.Join(ss[:3], "/"), true
		}
	} else if strings.HasPrefix(modPath, "git.inke.cn/tpc/inf") {
		ss := strings.Split(modPath, "/")
		if len(ss) >= 4 {
			return strings.Join(ss[:4], "/"), true
		}
	} else if strings.HasPrefix(modPath, "git.inke.cn/inkelogic/rpc-go") {
		ss := strings.Split(modPath, "/")
		if len(ss) >= 3 {
			return strings.Join(ss[:3], "/"), true
		}
	} else {
		r, err := realPath(modPath)
		if err != nil {
			return modPath, false
		}
		if len(r) > 0 {
			return r, true
		}
	}
	return modPath, false
}

func realPath(path string) (string, error) {
	reqUrl := "https://git.inke.cn/cache/" + path + "?go-get=1"
	realp := ""
	var err error
	defer func() {
		// doPrint("find real path from %s real: %s", reqUrl, realp)
		if err != nil {
			doError("find real path failed, from %s real: %s err: %q", reqUrl, realp, err)
		}
	}()

	tryAgain := false

	c := &http.Client{
		Timeout: 10 * time.Second,
	}
AGAIN:
	rsp, err := c.Get(reqUrl)
	if err != nil {
		doPrint("git cache go-get failed, uri:%s, error:%v", reqUrl, err)
		return "", err
	}

	if rsp.StatusCode != 200 {
		if !tryAgain {
			tryAgain = true
			rsp.Body.Close()
			reqUrl = "https://" + path + "?go-get=1"
			goto AGAIN
		}
		doPrint("git cache go-get failed, uri:%s, code:%d", reqUrl, rsp.StatusCode)
		return "", err
	}

	defer rsp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(rsp.Body)
	if err != nil {
		doPrint("goquery decode err %v", err)
		return "", err
	}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if v, ok := s.Attr("name"); !ok {
			return
		} else if v != "go-import" && v != "go-source" {
			return
		}
		v, ok := s.Attr("content")
		ss := strings.Split(v, " ")

		for _, s := range ss {
			if !strings.Contains(s, "/") {
				continue
			}
			if strings.HasPrefix(s, "git.inke.cn/cache/") &&
				strings.Contains(s, "github.com/golang") &&
				strings.HasPrefix(path, "golang.org/x") {

				s = strings.Replace(s, "github.com/golang", "golang.org/x", 1)
			}
			if strings.HasPrefix(s, "git.inke.cn/cache/") {
				s = strings.Replace(s, "git.inke.cn/cache/", "", 1)
			}

			if strings.Contains(path, strings.TrimSpace(s)) || strings.Contains(strings.TrimSpace(s), path) {
				realp = s
				break
			}
		}
		if ok && (len(realp) > 0) && (strings.Contains(path, strings.TrimSpace(realp)) || strings.Contains(strings.TrimSpace(realp), path)) {
			return
		}
		for _, s := range ss {
			if !strings.Contains(s, "/") {
				continue
			}
			if strings.Contains(s, "https://") {
				continue
			}
			realp = s
			break
		}
	})
	return strings.TrimSpace(realp), nil
}

func moveVendor(dir string) {
	_ = filepath.Walk(dir, func(filename string, fi os.FileInfo, err error) error { // 遍历目录
		if fi != nil && fi.IsDir() && strings.Contains(fi.Name(), "vendor") && !strings.Contains(fi.Name(), "vendor_bak") {
			if strings.Contains(filename, "git.inke.cn") {
				bak := filename + "_bak"
				// 移除基础库的vendor
				Move(filename, bak)
			}
		}
		return nil
	})
}
