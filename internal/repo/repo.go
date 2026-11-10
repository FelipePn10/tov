package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/FelipePn10/tov/internal/config"
)

type Repo struct {
	WorkTree string // raiz da arvore de trabalho ("" em repositorios bare)
	GitDir   string // o diretorio .git (ou a propia raiz se for um bare)
	Config   *config.Config
}

const DefaultBranch = "main"

// Cria um repositorio vazio em path, com bare nao ha arvore de trabalho o proprio path faz o papel de .git
// isso e o formato usado em servidores para onde faz push
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
	}
	for _, d := range []string{"objects/pack", "objects/info", "refs/heads", "refs/tags"} {
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
