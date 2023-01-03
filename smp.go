package gscp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type LoadOpts struct {
	path string
}

type Host struct {
	Name    string
	Options []Option
}

type Option struct {
	Name  string
	Value string
}

type option func(*LoadOpts)

func Path(p string) option {
	return func(l *LoadOpts) {
		l.path = p
	}
}

func LoadConfig(path ...option) (string, error) {

	home, _ := os.UserHomeDir()
	p := &LoadOpts{
		path: filepath.Join(home, ".ssh", "config"),
	}
	for _, opt := range path {
		opt(p)
	}
	f, err := os.Open(p.path)
	if err != nil {
		return "", fmt.Errorf("ERROR LoadConfig() Open: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("ERROR LoadConfig() ReadAll: %w", err)
	}
	return string(data), nil
}

func Parse(s string) ([]Host, error) {
	var hosts []Host
	reg := "\r\n|\n"
	tmp := regexp.MustCompile(reg).Split(s, -1)
	r := regexp.MustCompile(`^[a-zA-Z]`)
	for _, v := range tmp {
		if r.MatchString(v) {
			arr := strings.Fields(v)
			if strings.ToLower(arr[0]) == "host" {
				hosts = append(hosts, Host{Name: arr[1]})
			}
			if strings.ToLower(arr[0]) == "include" {
				home, _ := os.UserHomeDir()
				files, _ := filepath.Glob(filepath.Join(home, ".ssh", arr[1]))
				for _, f := range files {
					s, err := LoadConfig(Path(f))
					if err != nil {
						return nil, fmt.Errorf("ERROR Parse(): %w", err)
					}
					t, err := Parse(s)
					if err != nil {
						return nil, fmt.Errorf("ERROR Parse(): %w", err)
					}
					hosts = append(hosts, t...)
				}
			}
		} else {
			i := strings.TrimSpace(v)
			if len(i) == 0 {
				continue
			}
			if strings.HasPrefix(i, "#") {
				continue
			}
			s := strings.Fields(i)
			if len(s) == 0 {
				continue
			}
			if len(hosts) == 0 {
				continue
			}
			hosts[len(hosts)-1].Options = append(hosts[len(hosts)-1].Options, Option{Name: s[0], Value: s[1]})
		}
	}
	return hosts, nil
}
