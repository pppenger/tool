package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

var cmdBuild = &Command{
	Run:   runBuild,
	Name:  "build",
	Short: "该命令一般在编译机上运行,它根据Gedeps/Godeps.json文件中的依赖版本下载依赖.\n\t\t需要注意的是,本地执行的话,会将有变动的依赖包bak下,然后重新下载所需版本的依赖包,恢复的话需要执行recover命令.",
}

const (
	_cacheServiceName = "base.tool.package-cacher"
	_httpProto        = "http"
	_roundRobin       = "roundrobin"
	_consul           = "consul"
	_cacheReadTimeout = 3000
)

var _fallbackHosts = []string{"ali-f-ops-build02.bj:16270", "ali-e-ops-lark-runner01.bj:16270", "ali-f-ops-lark-runner03.bj:16270", "ali-g-ops-lark-runner02.bj:16270"}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func runBuild(args []string, ctx *Context) error {
	//需要自行的进行init,loadbalance
	// 1.去掉本地vendor
	if ok, _ := IsPathExist("vendor"); ok {
		moveVendor(ctx.WorkDir)
	}

	// 2.读取依赖文件
	godeps := &Godeps{}
	depFile := ctx.WorkDir + "/Godeps/Godeps.json"
	if err := godeps.OpenGodeps(depFile); err != nil {
		return err
	}
	// 灰度配置
	if yes, _ := IsPathExist(configFile); yes {
		buildMachine = true
		toml.DecodeFile(configFile, gConfig)
	}

	// 3.强制保证基础库必须存在,默认基础库必须用master分支
	for baseMod := range necessaryDeps {
		found := false
		for _, g := range godeps.Deps {
			if strings.Contains(g.ImportPath, baseMod) {
				found = true
				break
			}
		}
		if !found {
			cc := CustomConfig{SpecificTag: "master", SpecificRev: ""}
			dep := &Dependence{
				ImportPath: baseMod,
				Custom:     cc,
			}
			godeps.Deps = append(godeps.Deps, dep)
		}
	}

	// 4.下载依赖
	s1 := time.Now()
	if strings.HasPrefix(strings.TrimSpace(godeps.InkedepVersion), "2.0") {
		buildv2(ctx, godeps)
	} else {
		buildv1(ctx, godeps)
	}
	doPrint("--------deps file build end, cost: %.3fs---------", time.Now().Sub(s1).Seconds())

	s1 = time.Now()
	// 5.再次扫描依赖,补全可能缺失的依赖
	root, err := FindGitRoot(ctx.WorkDir)
	if err != nil {
		return nil
	}

	path := filepath.Join(ctx.GOPATH, "src")
	dir, err := filepath.Rel(path, root)
	if err != nil {
		return nil
	}

	times := 0
Again:
	doPrint("--------start rescan deps---------")
	hadScaned = sync.Map{}
	depsOnBuilding := ctx.loadDeps(root, dir, true)
	for mod := range depsOnBuilding {
		found := false
		for m := range finalDepdences {
			if strings.Contains(mod, m) || strings.Contains(m, mod) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		moddir := ctx.GopathSrc(mod)
		if ok, _ := IsPathExist(moddir); !ok {
			doPrint(">>>> doing reload, maybe missing dep, mod: %s", mod)
			Remove(moddir)
			runGet([]string{mod}, ctx)
			continue
		}
		rev, _ := vcsGit.Identify(moddir)
		if len(rev) > 0 {
			w := &whoUseDep{
				revision: rev[:7],
			}
			finalDepdences[mod] = w
		}
	}
	times += 1
	if times < 3 {
		goto Again
	}

	doPrint("--------reload deps end, cost: %.3fs---------", time.Now().Sub(s1).Seconds())

	// 6.递归移除gopath/src下的vendor目录
	// moveVendor(ctx.GopathSrc(""))

	// 上报基础库版本信息
	if buildMachine {
		postBaseDesps(dir, ctx.WorkDir)
	}

	// 7.保存依赖
	writeDeps(ctx)

	return nil
}

const pat = `_cluster\.([^_\./]+)`

var reg = regexp.MustCompile(pat)

func postBaseDesps(thisMod string, workDir string) {
	meta := DepsMeta{}
	meta.Service = thisMod

	subs := reg.FindStringSubmatch(workDir)
	if len(subs) >= 2 {
		meta.Cluster = subs[1]
	}

	for mod, d := range finalDepdences {
		if strings.Contains(mod, "git.inke.cn/inkelogic/daenerys") {
			meta.Dae = d.revision
		} else if strings.Contains(mod, "git.inke.cn/inkelogic/rpc-go") {
			meta.RpcGo = d.revision
		} else if strings.Contains(mod, "git.inke.cn/tpc/inf/go-upstream") {
			meta.Upstream = d.revision
		} else if strings.Contains(mod, "git.inke.cn/BackendPlatform/golang") {
			meta.Golang = d.revision
		} else if strings.Contains(mod, "git.inke.cn/tpc/inf/metrics") {
			meta.Metric = d.revision
		} else if strings.Contains(mod, "git.inke.cn/BackendPlatform/jaeger-client-go") {
			meta.Jaeger = d.revision
		}
	}
	b, _ := json.Marshal(meta)
	http.Post("http://10.111.174.25:55555/inkedep", "application/json; charset=utf-8", bytes.NewBuffer(b))
}

var recovering = map[string]string{}
var finalDepdences = map[string]*whoUseDep{}
var aLock sync.Mutex

type whoUseDep struct {
	tag      string
	revision string
	whoUsing []string
}

func writeDeps(ctx *Context) {
	releasePath := ctx.WorkDir
	if strings.HasSuffix(ctx.WorkDir, "src") {
		ss := strings.Split(releasePath, "/")
		releasePath = strings.Join(ss[:len(ss)-1], "/")
	}
	depsFile := ".deps"
	depsFilePath := releasePath + "/release/" + depsFile
	f, e := os.OpenFile(depsFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if e != nil {
		doPrint(">>>> create .deps file fail, err: %v", e)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for mod, dep := range finalDepdences {
		s := fmt.Sprintf("revision: %s module: %s\n", dep.revision, mod)
		w.Write([]byte(s))
	}
	w.Flush()
	doPrint(depsFilePath)
}

func isTag(dir string, tag string) bool {
	t := tag
	ss := strings.Split(tag, "-")
	if len(ss) > 0 {
		t = ss[0]
	}
	tt, _ := vcsGit.ListTags(dir)
	if strings.Contains(tt, t) {
		return true
	}
	return false
}

func specialIgnored(beUsedMod, modUser string) bool {
	if strings.Contains(beUsedMod, "git.inke.cn/BackendPlatform/golang") ||
		strings.Contains(beUsedMod, "git.inke.cn/inkelogic/rpc-go") {

		if strings.Contains(modUser, "git.inke.cn/BackendPlatform/golang") ||
			strings.Contains(modUser, "git.inke.cn/inkelogic/rpc-go") {
			return true
		}
	}
	return false
}

func download(ctx *Context, crt *Concurrent, godeps *Godeps) {
	modulePath := godeps.ImportPath
	for _, dep := range godeps.Deps {
		if ok, _ := ctx.buildInPackage(dep.ImportPath); ok {
			continue
		}
		if strings.Contains(ctx.WorkDir, dep.ImportPath) {
			continue
		}
		dep := dep
		// 灰度设置
		setGrayConfig(ctx.WorkDir, dep)

		// 已经缓存过了,则忽略
		aLock.Lock()
		if m, ok := finalDepdences[dep.ImportPath]; ok {
			if dep.Custom.SpecificTag == m.tag && len(dep.Custom.SpecificRev) == 0 {
				aLock.Unlock()
				continue
			}
			if strings.Contains(dep.Custom.SpecificRev, m.revision) || strings.Contains(m.revision, dep.Custom.SpecificRev) {
				aLock.Unlock()
				continue
			}
		}
		aLock.Unlock()

		closure := func() (interface{}, error) {
			gitCache := false
			commitCache := false
			dir := ctx.GopathSrc(dep.ImportPath)
			defer func() {
				var rev string
				var err error
				var branch string
				if rev, err = vcsGit.Identify(dir); err == nil {
					branch, _ = vcsGit.Branch(dir)
					aLock.Lock()
					if value, ok := finalDepdences[dep.ImportPath]; !ok {
						user := make([]string, 0, 5)
						w := &whoUseDep{
							tag:      dep.Custom.SpecificTag,
							revision: rev[:7],
							whoUsing: append(user, modulePath),
						}
						finalDepdences[dep.ImportPath] = w
					} else {
						user := value.whoUsing
						user = append(user, dep.ImportPath)
						w := &whoUseDep{
							tag:      dep.Custom.SpecificTag,
							revision: rev[:7],
							whoUsing: user,
						}
						finalDepdences[dep.ImportPath] = w
					}
					aLock.Unlock()
				}
				if commitCache {
					return
				}
				prev := rev
				if len(rev) > 0 {
					prev = rev[:7]
				}
				doPrint("cached:%v, rev:%s\t tag:%-20s\t%s", gitCache, prev, branch, dep.ImportPath)
			}()

			var dstTag string
			if ok, _ := IsPathExist(dir); ok {
				// repo: git@xxx 或者https://xxx
				repo, _, err := FindRepoCacheRoot(dep.changePath())
				if err == nil {
					// 修改本地代码关联的远程地址
					err = vcsGit.Remote(dir, repo)
					if err != nil {
						doPrint(">>>> set remote repo: %s fail, path: %s, err: %q", repo, dep.ImportPath, err)
						goto FROMREMOTE
					}
				}

				// 非编译机执行build
				if !buildMachine {
					curBranch, _ := vcsGit.Branch(dir)
					curRev, _ := vcsGit.Identify(dir)
					if len(dep.Custom.SpecificTag) == 0 {
						doError("dep config invalid, module: %+v", *dep)
						return nil, nil
					}

					if len(curBranch) > 0 && curBranch != dep.Custom.SpecificTag {
						doRecover(ctx, dir, dep)
						goto FROMREMOTE
					}

					if len(dep.Custom.SpecificRev) > 0 {
						if len(curRev) > 0 && curRev != dep.Custom.SpecificRev {
							doRecover(ctx, dir, dep)
							goto FROMREMOTE
						}
					}
				}

				specific := false // 是否指定了tag/版本
				if len(dep.Custom.SpecificTag) > 0 {
					dstTag = dep.Custom.SpecificTag
					yes := isTag(dir, dstTag)
					// RevSync: git checkout <tag/branch>
					e := vcsGit.RevSync(dir, dstTag) // 切到指定分支或tag
					if e == nil {
						specific = true
						gitCache = true
					}
					if specific && !yes { //指定了分支名而非tag,需要同步最新代码
						// git pull,获取最新代码
						e1 := vcsGit.Update(dir)
						if e1 != nil {
							goto FROMREMOTE
						}
					}
					// branch和rev同时指定
					if !yes && len(dep.Custom.SpecificRev) > 0 {
						e = vcsGit.RevSync(dir, dep.Custom.SpecificRev) // 切到指定commit
						if e != nil {
							specific = false
						}
					}
				} else if len(dep.Custom.SpecificRev) > 0 {
					e := vcsGit.RevSync(dir, dep.Custom.SpecificRev) // 切到指定commit
					if e == nil {
						specific = true
						gitCache = true
					}
				}

				// 切到指定tag失败了
				if !specific && (len(dep.Custom.SpecificTag) > 0 || len(dep.Custom.SpecificRev) > 0) {
					doError("\033[43;34m can't checkout, module %s, tag: %s \033[0m", dep.ImportPath, dep.Custom.SpecificTag)
					goto FROMREMOTE
				} else {
					aLock.Lock()
					latestRev, _ := vcsGit.Identify(dir)
					if cached, ok := finalDepdences[dep.ImportPath]; ok {
						users := cached.whoUsing
						cacheRev := cached.revision
						if !specialIgnored(dep.ImportPath, modulePath) &&
							!strings.HasPrefix(latestRev, cacheRev) { // 同一个包,被引用了不同版本,忽略使用缓存中的该包,因为缓存中的版本可能不是最新的
							doError("\033[41;30m rev conflict, module: %s, %v-dep:%s, [%s]-dep:%s \033[0m", dep.ImportPath, users, cacheRev, modulePath, latestRev)
							aLock.Unlock()
							os.Exit(-1)
						}
						commitCache = true
					}
					aLock.Unlock()
					return nil, nil
				}
			}

		FROMREMOTE:
			if e := Remove(dir); e != nil {
				return nil, e
			}
			aLock.Lock()
			defer aLock.Unlock()
			t2 := time.Now()
			// git clone and checkout
			d := ctx.GopathSrc(dep.ImportPath)
			if err := dep.Create(d); err != nil {
				doError("on build stage, create fail, err:%v, dep:%+v\n", err, *dep)
				return nil, err
			}
			vals := make(url.Values)
			vals.Add("repo", dep.ImportPath)
			vals.Add("work_dir", ctx.WorkDir)
			go http.Get(fmt.Sprintf("http://%s/api/v1/package/cache?", _fallbackHosts[rand.Intn(len(_fallbackHosts))]) + vals.Encode())
			doPrint("download cost:[%.3fs], package:%s", time.Now().Sub(t2).Seconds(), dep.ImportPath)
			return nil, nil
		}
		crt.Do(closure)
	}
}

func buildv1(ctx *Context, godeps *Godeps) {
	modulePath := godeps.ImportPath
	crt := NewConcurrent()
	for _, dep := range godeps.Deps {
		if ok, _ := ctx.buildInPackage(dep.ImportPath); ok {
			continue
		}
		if strings.Contains(ctx.WorkDir, dep.ImportPath) {
			continue
		}
		dep := dep
		// 灰度设置
		setGrayConfig(ctx.WorkDir, dep)

		if baseLib(dep.ImportPath, baseModulePath) || baseLib(dep.ImportPath, baseModulePath2) {
			if err := downloadBaseLib(ctx, crt, dep.ImportPath); err != nil {
				doError(">>>> on v1, download baselib's deps failed, error %q", err)
				return
			}
		}

		// download
		closure := func() (interface{}, error) {
			localCached := false
			commitCached := false
			dir := ctx.GopathSrc(dep.ImportPath)
			defer func() {
				var rev string
				var err error
				if rev, err = vcsGit.Identify(dir); err == nil {
					aLock.Lock()
					if value, ok := finalDepdences[dep.ImportPath]; !ok {
						user := make([]string, 0, 5)
						w := &whoUseDep{
							tag:      dep.Rev,
							revision: rev[:7],
							whoUsing: append(user, modulePath),
						}
						finalDepdences[dep.ImportPath] = w
					} else {
						user := value.whoUsing
						user = append(user, dep.ImportPath)
						w := &whoUseDep{
							tag:      dep.Rev,
							revision: rev[:7],
							whoUsing: user,
						}
						finalDepdences[dep.ImportPath] = w
					}
					aLock.Unlock()
				}
				if commitCached {
					return
				}
				doPrint("on v1, cached:%v, rev:%s %s", localCached, dep.Rev, dep.ImportPath)
			}()

			if ok, _ := IsPathExist(dir); ok {
				if rev, err := vcsGit.Identify(dir); err == nil && rev == dep.Rev {
					localCached = true
					return nil, nil
				}
				repo, _, err := FindRepoCacheRoot(dep.changePath())
				if err == nil {
					err = vcsGit.Remote(dir, repo)
					if err != nil {
						return nil, err
					}
				}
				e2 := vcsGit.RevSync(dir, dep.Rev)
				e1 := vcsGit.Update(dir)
				if e1 == nil && e2 == nil {
					localCached = true
					aLock.Lock()
					if _, ok := finalDepdences[dep.ImportPath]; ok {
						commitCached = true
					}
					aLock.Unlock()
					return nil, nil
				} else {
					doPrint(">>>> on v1, checkout to %s failed, or pull code failed, package %s\n", dep.Rev, dep.ImportPath)
				}

				if e := Remove(dir); e != nil {
					return nil, e
				}
			}

			aLock.Lock()
			defer aLock.Unlock()
			t2 := time.Now()
			d := ctx.GopathSrc(dep.ImportPath)
			if err := dep.Create(d); err != nil {
				doError(">>>> on v1, build stage, create fail, err:%v, dep:%+v\n", err, *dep)
				return nil, err
			}
			vals := make(url.Values)
			vals.Add("repo", dep.ImportPath)
			vals.Add("work_dir", ctx.WorkDir)
			go http.Get(fmt.Sprintf("http://%s/api/v1/package/cache?", _fallbackHosts[rand.Intn(len(_fallbackHosts))]) + vals.Encode())
			doPrint("on v1, download cost:[%.3fs], package:%s", time.Now().Sub(t2).Seconds(), dep.ImportPath)
			return nil, nil
		}
		crt.Do(closure)
	}
	_, err := crt.DoDone()
	if err != nil {
		doError(">>>> on v1, download direct deps failed, error %q", err)
		return
	}
	return
}

func downloadBaseLib(ctx *Context, crt *Concurrent, module string) error {
	doPrint(">>>> download deps for baseLib, module: %s", module)
	dir := filepath.Join(ctx.GOPATH, "src", module)
	if ok, _ := IsPathExist(dir); !ok {
		runGet([]string{module}, ctx)
	}
	basePath := dir + "/Godeps/Godeps.json"
	baseDeps := &Godeps{}
	if err := baseDeps.OpenGodeps(basePath); err != nil {
		doPrint(">>>> open deps file failed: %v", err)
		return err
	}
	download(ctx, crt, baseDeps)
	return nil
}

func buildv2(ctx *Context, godeps *Godeps) {
	for _, g := range godeps.Deps {
		if len(g.Custom.SpecificTag) == 0 && len(g.Custom.SpecificRev) == 0 {
			doError(">>>> dep config err, mod: %s", g.ImportPath)
			return
		}
	}

	doPrint("---------------------start direct deps %d---------------------\n", len(godeps.Deps))
	crt := NewConcurrent()
	download(ctx, crt, godeps)
	_, err := crt.DoDone()
	if err != nil {
		doError(">>>> download direct deps failed, error %q", err)
		return
	}

	doPrint("---------------------start baselib's deps-------------------------\n\n")
	// 下载基础库的依赖包
	crtBase := NewConcurrent()
	for _, d := range godeps.Deps {
		dep := d
		if baseLib(dep.ImportPath, baseModulePath) || baseLib(dep.ImportPath, baseModulePath2) {
			if err := downloadBaseLib(ctx, crtBase, dep.ImportPath); err != nil {
				doError(">>>> download baselib deps failed, error %q", err)
				return
			}
		}
	}
	_, err = crtBase.DoDone()
	if err != nil {
		doError(">>>> download direct deps failed, error %q", err)
		return
	}
	return
}

func onGraying(workDir string) bool {

	// exclude
	for _, v := range gConfig.Exclude {
		if strings.Contains(workDir, v) {
			return false
		}
	}

	// job name
	for _, v := range gConfig.Jobs {
		if strings.Contains(workDir, v) {
			return true
		}
	}

	// node config
	for _, v := range gConfig.ServiceTree {
		var owt, pdl, servicegroup, cluster string
		if len(v.Owt) > 0 {
			owt = "_owt." + v.Owt
		}
		if len(v.Pdl) > 0 {
			pdl = "_pdl." + v.Pdl
		}
		if len(v.ServiceGroup) > 0 {
			servicegroup = "_servicegroup." + v.ServiceGroup
		}
		if len(v.Cluster) > 0 {
			cluster = "_cluster." + v.Cluster
		}

		if len(owt) > 0 && len(pdl) > 0 {
			tag := owt + pdl
			if len(servicegroup) > 0 {
				tag += servicegroup
			}
			if strings.Contains(workDir, tag) {
				if len(cluster) > 0 && !strings.Contains(workDir, cluster) {
					continue
				}
				return true
			}
		}

		if len(owt) > 0 {
			if strings.Contains(workDir, owt) {
				if len(cluster) > 0 && !strings.Contains(workDir, cluster) {
					continue
				}
				return true
			}
		}
	}

	// cluster
	if gConfig.Cluster.Open {
		for _, v := range gConfig.Cluster.List {
			if len(v) > 0 {
				cluster := "_cluster." + v
				if strings.Contains(workDir, cluster) {
					return true
				}
			}
		}
	}
	return false
}

//func dingdingRobot(fn, url, workDir, module, rev string) {
//	printable := true
//	if strings.Contains(workDir, "test") && strings.Contains(workDir, "_cluster") {
//		printable = false
//	}
//	if printable {
//		job := fmt.Sprintf("基础库模块: %s\n基础库版本: %s\n任务名: %s", module, rev, workDir)
//		body := fmt.Sprintf(`{"msgtype": "text", "text": {"content":"%s"}}`, job)
//		http.Post(url, "application/json", bytes.NewBufferString(body))
//		if len(fn) == 0 {
//			return
//		}
//		f, e := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
//		if e != nil {
//			return
//		}
//		defer f.Close()
//		f.WriteString(body)
//	}
//}

func setGrayConfig(workDir string, dep *Dependence) {
	if baseLib(dep.ImportPath, necessaryDeps) && onGraying(workDir) {
		for _, info := range gConfig.Customs {
			if strings.Contains(dep.ImportPath, info.Dep) ||
				strings.Contains(info.Dep, dep.ImportPath) {
				if len(info.Tag) > 0 {
					dep.Custom.SpecificTag = info.Tag
					dep.Custom.SpecificRev = info.Tag
					dep.Rev = info.Tag // 兼容老版本
				} else if len(info.Tag) == 0 && len(info.Rev) > 0 {
					dep.Custom.SpecificRev = info.Rev
					dep.Rev = info.Rev // 兼容老版本
				}
				if gConfig.DDRobot.Open {
					// dingdingRobot(gConfig.GrayLog, gConfig.DDRobot.Url, workDir, dep.ImportPath, dep.Rev)
				}
				break
			}
		}
	}

	if strings.Contains(dep.ImportPath, "github.com/robfig/cron") && dep.Custom.SpecificTag == "master" {
		dep.Custom.SpecificTag = "v1"
	}
}

func doRecover(ctx *Context, dir string, dep *Dependence) {
	// mv dir dir.bak
	bak := fmt.Sprintf("%s.bak", dir)
	if ok, _ := IsPathExist(bak); ok {
		return
	}
	Move(dir, bak)
	recovering[dir] = bak
	// git clone dir
	d := ctx.GopathSrc(dep.ImportPath)
	if err := dep.Create(d); err != nil {
		doWarning("on doRebuild, create err:%q, dep:%+v\n", err, dep)
	}
	// todo: finally, rm -rf dir, mv dir.bak dir
}
