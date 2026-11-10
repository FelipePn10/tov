//// [core]
// 	repositoryformatversion = 0
// 	filemode = true
// 	bare = false
// [remote "origin"]
// 	url = https://github.com/exemplo/projeto.git
// 	fetch = +refs/heads/*:refs/remotes/origin/*
// [user]
// 	name = Ana Lima
// 	email = ana@exemplo.com

package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
)

type entry struct{ key, value string }

type section struct {
	name, sub string
	entries   []entry
}

// Config mantém as seções na ordem do arquivo, para regravá-lo sem
// embaralhar o que o usuário escreveu.
type Config struct {
	path     string
	sections []*section
}

// Le o arquivo, se nao tiver vai devolver uma configuraçao vazia
func Load(path string) (*Config, error) {
	c := &Config{path: path}
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var cur *section
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' || line[0] == ';' {
			continue
		}
		if line[0] == '[' {
			end := strings.LastIndexByte(line, ']')
			if end < 0 {
				return nil, fmt.Errorf("%s:%d: section sem ']'", path, n)
			}
			name, sub, _ := strings.Cut(line[1:end], " ")
			cur = &section{name: strings.ToLower(name), sub: strings.Trim(strings.TrimSpace(sub), `"`)}
			c.sections = append(c.sections, cur)
			continue
		}
		if cur == nil {
			return nil, fmt.Errorf("%s:%d: chave fora de seçao", path, n)
		}
		k, v, found := strings.Cut(line, "=")
		if !found {
			v = "true" // "[core] bare" sozinho significa bare = true
		}
		cur.entries = append(cur.entries, entry{
			key:   strings.ToLower(strings.TrimSpace(k)),
			value: strings.Trim(strings.TrimSpace(v), `"`),
		})
	}
	return c, sc.Err()
}

func (c *Config) find(name, sub string) *section {
	name = strings.ToLower(name)
	for _, s := range c.sections {
		if s.name == name && s.sub == sub {
			return s
		}
	}
	return nil
}

// Get devolve o valor de section.sub.key (vazio para seçoes simples)
// Se a chave aparece varias vezes, vale a ultima, assim como e no Get
func (c *Config) Get(name, sub, key string) (string, bool) {
	name = strings.ToLower(name)
	val, ok := "", false
	for _, s := range c.sections {
		if s.name != name || s.sub != sub {
			continue
		}
		for _, e := range s.entries {
			if e.key == strings.ToLower(key) {
				val, ok = e.value, true
			}
		}
	}
	return val, ok
}

// Define section.sub.key
func (c *Config) Set(name, sub, key, value string) {
	key = strings.ToLower(key)
	s := c.find(name, sub)
	if s == nil {
		s = &section{name: name, sub: sub}
		c.sections = append(c.sections, s)
	}
	// Percorrer apartir do ultimo e nao do primeiro, pois quero que faça sentido com o metodo Get, afinal o Get atualiza o ultimo tbm
	for i, e := range slices.Backward(s.entries) {
		if e.key == key {
			s.entries[i].value = value
			return
		}
	}
	s.entries = append(s.entries, entry{key, value})
}

// Lista os subnomes de uma seçao (ex.: nome dos remotes)
func (c *Config) Subsections(name string) []string {
	var out []string
	for _, s := range c.sections {
		if s.name == name && s.sub != "" {
			out = append(out, s.sub)
		}
	}
	return out
}

// Apaga uma seçao inteira
func (c *Config) RemoveSection(name, sub string) bool {
	for i, s := range c.sections {
		if s.name == name && s.sub == sub {
			c.sections = append(c.sections[:i], c.sections[i+1:]...)
			return true
		}
	}
	return false
}

// save grava o arquivo. (Versao provisoria ok) depois vou voltar e adc um pacote lockfile para gravar de alguma forma atomica e com trava
func (c *Config) Save() error {
	var b strings.Builder
	for _, s := range c.sections {
		if s.sub != "" {
			fmt.Fprintf(&b, "[%s %q]\n", s.name, s.sub)
		} else {
			fmt.Fprintf(&b, "[%s]\n", s.name)
		}
		for _, e := range s.entries {
			fmt.Fprintf(&b, "\t%s = %s\n", e.key, e.value)
		}
	}
	return os.WriteFile(c.path, []byte(b.String()), 0o644)
}
