package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func main() {
	targets := new(Targets)
	format := flag.String("name", "", "Ignore mobile platforms (Android, iOS)")
	mobile := flag.Bool("ignore-mobile", false, "Ignore mobile platforms (Android, iOS)")
	web := flag.Bool("ignore-web", false, "Ignore web platforms (WebAssembly)")
	flag.Parse()

	if web != nil && *web {
		targets.Filter("js")
	}

	if mobile != nil && *mobile {
		targets.Filter("android")
		targets.Filter("ios")
	}

	if format == nil || strings.TrimSpace(*format) == "" {
		fmt.Println("[goall] specify -name, it should be something like '-name something_{{OS}}_{{ARCH}}'.")
		os.Exit(1)
		return
	}

	cpus := runtime.NumCPU()
	if cpus < 4 {
		cpus = 4
	}

	builders := NewBuilders(cpus/4, flag.Args(), *format)
	for sys, arch, ok := targets.Next(); ok; sys, arch, ok = targets.Next() {
		builders.Build(sys, arch)
	}
	builders.Wait()
}

type Builders struct {
	args   []string
	orders chan [2]string
	group  *sync.WaitGroup
}

func (b *Builders) Build(os, arch string) {
	b.orders <- [2]string{os, arch}
}

func (b *Builders) Wait() {
	close(b.orders)
	b.group.Wait()
}

func NewBuilders(n int, args []string, format string) *Builders {
	b := Builders{args: args, orders: make(chan [2]string, n), group: new(sync.WaitGroup)}
	for i := 0; i < n; i++ {
		b.group.Add(1)
		go func() {
			defer b.group.Done()
			for v := range b.orders {

				name := format
				name = strings.Replace(name, `{{OS}}`, v[0], -1)
				name = strings.Replace(name, `{{ARCH}}`, v[1], -1)
				if v[0] == "windows" {
					name += ".exe"
				}

				args := make([]string, len(b.args)-1)
				copy(args, b.args[:len(args)])
				isOutputSet := false
				for i, a := range args {
					if a == "-o" && i != len(args)-1 {
						args[i+1] = filepath.Join(args[i+1], name)
						isOutputSet = true
					}
				}
				if !isOutputSet {
					args = append(args, []string{`-o`, name}...)
				}
				args = append(args, b.args[len(b.args)-1])

				cmd := exec.Command(`go`, args...)
				cmd.Env = append(os.Environ(), "GOOS="+v[0], "GOARCH="+v[1])
				if result, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf(`[goall][%s/%s] error compiling: %v`, v[0], v[1], err)
					fmt.Printf(`[goall][%s/%s] error compiling: %v`, v[0], v[1], string(result))
				}
			}
		}()
	}

	return &b
}

type Targets struct {
	filters map[string]struct{}
	list    []byte
	last    int
}

func (t *Targets) Filter(s string) {
	if t.filters == nil {
		t.filters = make(map[string]struct{}, 8)
	}
	t.filters[s] = struct{}{}
}

func (t *Targets) Next() (os, arch string, ok bool) {
	if t.list == nil {
		list, err := exec.Command(`go`, `tool`, `dist`, `list`).CombinedOutput()
		if err != nil {
			return os, arch, false
		}
		t.list = list
	}

	for i := t.last; i < len(t.list); i++ {
		if t.list[i] == '/' {
			os = string(t.list[t.last:i])
			t.last = i + 1
		}
		if t.list[i] == '\n' || i == len(t.list) {
			arch = string(t.list[t.last:i])
			t.last = i + 1
			if _, found := t.filters[os]; !found {
				return os, arch, true
			}
		}
	}
	return os, arch, false
}
