// Command scaffold copies this template into a fresh project, rewriting the
// module path and application name everywhere. The scaffold command itself is
// excluded from the copy, so generated projects ship only the application.
//
// Usage:
//
//	go run ./cmd/scaffold -module github.com/acme/blog -out ../blog [-name blog] [-src .]
//
// Then in the target directory:
//
//	go mod tidy && go build ./... && go run ./cmd/api
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	oldModule   = "github.com/Wi1low/go-web-template"
	oldAppShort = "go-web-template"
	// Placeholder shields the injected module path from the short-name pass:
	// a new module like github.com/acme/go-web-template-kit must not be
	// rewritten out of recognition.
	modulePlaceHolder = "\x00MODULE_PATH\x00"
)

// Directories never copied into a new project.
var skipDirs = map[string]bool{
	".git": true, "bin": true, "logs": true,
	"cmd" + string(filepath.Separator) + "scaffold": true,
}

// Extensions (plus exact names) treated as text and rewritten. Everything
// else is copied byte-for-byte.
var textExt = map[string]bool{
	".go": true, ".mod": true, ".sum": true, ".yaml": true, ".yml": true,
	".md": true, ".json": true, ".txt": true, ".tmpl": true, ".html": true,
}
var textNames = map[string]bool{
	"Makefile": true, "Dockerfile": true, ".env.example": true, ".gitignore": true,
}

func main() {
	var (
		module = flag.String("module", "", "new Go module path, e.g. github.com/acme/blog (required)")
		out    = flag.String("out", "", "target directory (required, must not exist or be empty)")
		name   = flag.String("name", "", "application short name (default: last segment of -module)")
		src    = flag.String("src", ".", "template source root (default: current directory)")
	)
	flag.Parse()

	if err := run(*module, *name, *out, *src); err != nil {
		fmt.Fprintln(os.Stderr, "scaffold:", err)
		os.Exit(1)
	}
}

func run(module, name, out, src string) error {
	if module == "" || out == "" {
		return errors.New("both -module and -out are required")
	}
	if name == "" {
		name = module[strings.LastIndexAny(module, "/:")+1:]
	}
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(absSrc, absOut); err == nil && !strings.HasPrefix(rel, "..") {
		return fmt.Errorf("-out %s must be outside the template source %s", absOut, absSrc)
	}
	if err := os.MkdirAll(absOut, 0o755); err != nil {
		return err
	}

	count := 0
	err = filepath.WalkDir(absSrc, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if d.IsDir() {
			if skipDirs[rel] {
				return fs.SkipDir
			}
			return os.MkdirAll(filepath.Join(absOut, rel), 0o755)
		}
		if rel == ".env" { // secrets never travel
			return nil
		}
		if err := copyFile(filepath.Join(absOut, rel), path, module, name); err != nil {
			return fmt.Errorf("copy %s: %w", rel, err)
		}
		count++
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("project %q created at %s (%d files)\n", name, absOut, count)
	fmt.Println("next steps:")
	fmt.Println("  cd " + filepath.ToSlash(out))
	fmt.Println("  go mod tidy && go build ./...")
	fmt.Println("  go run ./cmd/api        # boots with dev config, Ctrl+C stops cleanly")
	return nil
}

func copyFile(dst, src, module, name string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if !isTextFile(filepath.Base(src)) {
		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	// Underscore form of the app name, used by the metric namespace and the
	// demo database name. A new name like "my-app" becomes "my_app".
	nameUnderscore := strings.NewReplacer("-", "_", ".", "_").Replace(name)

	s := string(data)
	s = strings.ReplaceAll(s, oldModule, modulePlaceHolder)
	s = strings.ReplaceAll(s, oldAppShort, name)
	s = strings.ReplaceAll(s, modulePlaceHolder, module)
	// Config-level identifiers that follow the underscore convention.
	s = strings.ReplaceAll(s, "go_web_template", nameUnderscore)
	s = strings.ReplaceAll(s, "web_template", nameUnderscore)
	return os.WriteFile(dst, []byte(s), 0o644)
}

func isTextFile(base string) bool {
	if textNames[base] {
		return true
	}
	return textExt[filepath.Ext(base)]
}
