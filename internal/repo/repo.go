package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/FelipePn10/tov/internal/config"
)

type Repo struct {
	WorkTree string // ("" em repositorios bare)
	GitDir   string // o diretorio .git (ou a propia raiz se for um bare)
	Config   *config.Config
}

const DefaultBranch = "main"

// Cria um repositorio vazio em path, com bare nao ha arvore de trabalho o proprio path faz o papel de .git
func Init(path string, bare bool) (*Repo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	gitDir := filepath.Join(abs, ".git")
	if bare {
		gitDir = abs
	}
	if _, err := os.Stat(filepath.Join(gitDir, "HEAD")); err == nil {
		return nil, fmt.Errorf("já existe um repositório em %s", gitDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("não foi possível acessar %s: %w", gitDir, err)
	}
	for _, d := range []string{
		"objects/pack",
		"objects/info",
		"refs/heads",
		"refs/tags",
	} {
		if err := os.MkdirAll(filepath.Join(gitDir, d), 0o755); err != nil {
			return nil, err
		}
	}
	head := []byte("ref: refs/heads/" + DefaultBranch + "\n")
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), head, 0o644); err != nil {
		return nil, err
	}
	cfg, err := config.Load(filepath.Join(gitDir, "config"))
	if err != nil {
		return nil, err
	}
	cfg.Set("core", "", "repositoryformatversion", "0")
	cfg.Set("core", "", "filemode", "true")
	cfg.Set("core", "", "bare", strconv.FormatBool(bare))
	if err := cfg.Save(); err != nil {
		return nil, err
	}
	return openAt(abs, gitDir, bare)
}

// Sob os diretorios ate encontrar um .git (basicamente como o git funciona dentro das subpastas)
func Open(start string) (*Repo, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	for {
		if isGitDir(filepath.Join(dir, ".git")) {
			return openAt(dir, filepath.Join(dir, ".git"), false)
		}
		if isGitDir(dir) { // bare
			return openAt(dir, dir, true)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, errors.New("não é um repositório (nem em nenhum diretório acima); use `kit init`")
		}
		dir = parent
	}
}

func (r *Repo) Path(elem ...string) string {
	return filepath.Join(append([]string{r.GitDir}, elem...)...)
}

func isGitDir(dir string) bool {
	entries := []struct {
		name  string
		isDir bool
	}{
		{"HEAD", false},
		{"objects", true},
		{"refs", true},
	}

	for _, entry := range entries {
		info, err := os.Stat(filepath.Join(dir, entry.name))
		if err != nil {
			return false
		}

		if info.IsDir() != entry.isDir {
			return false
		}
	}

	return true
}

func openAt(root, gitDir string, bare bool) (*Repo, error) {
	cfg, err := config.Load(filepath.Join(gitDir, "config"))
	if err != nil {
		return nil, err
	}
	r := &Repo{GitDir: gitDir, Config: cfg}
	if !bare {
		r.WorkTree = root
	}
	return r, nil
}
