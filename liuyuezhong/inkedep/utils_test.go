package main

import (
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	t.Log(FindRepoRoot("git.inke.cn/github/olivere/elastic"))
	t.Log(FindRepoRoot("code.inke.cn/bpc/server/search/listserv/list_server/pkg"))
	t.Log(FindRepoRoot("git.inke.cn/bpc/server/search/listserv/list_server/pkg"))
	t.Log(FindRepoRoot("git.inke.cn/bpc/server/search/listserv/list_server"))
	t.Log(FindRepoRoot("git.apache.org/thrift.git/lib/go/thrift"))
	//	t.Log(FindRepoRoot("golang.org/x/tools/go/vcs"))
	t.Log(FindRepoRoot("github.com/golang/tools/go/vcs"))
	t.Log(FindRepoRoot("git.inke.cn/inkelogic/rpc-go"))
	t.Log(FindRepoRoot("gopkg.in/go-playground/validator.v8"))
	t.Log(FindRepoRoot("go.uber.org/zap"))
	t.Log(FindRepoRoot("launchpad.net/gozk"))
	t.Log(FindRepoRoot("git.inke.cn/bpc/server/user/deploy/national_regulation/src"))
}
