package main

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInline(t *testing.T) {
	// Копируем библиотеку один раз для всех подтестов
	libDir := t.TempDir()
	if err := copyLibrary(libDir); err != nil {
		t.Fatal(err)
	}
	mustRun(t, libDir, "go", "mod", "tidy")

	tests := []struct {
		name       string
		content    string
		buildTags  string
		wantInline bool // true если инлайн должен успешно выполниться
		wantClear  bool // true если очистка должна успешно выполниться (после успешного инлайна)
		skipClear  bool // не проверять очистку
	}{
		{
			name: "dot_import",
			content: `package main
import (
    . "github.com/aaa2ppp/contestio"
    "os"
)
func main() {
    br := NewReader(os.Stdin)
    bw := NewWriter(os.Stdout)
    defer bw.Flush()
    var n int
    ScanIntLn(br, &n)
    a := make([]int, n)
    ScanInts(br, a)
    PrintIntsLn(bw, a)
}`,
			wantInline: true,
			wantClear:  true,
		},
		{
			name: "with_sugar_tag",
			content: `package main
import (
    . "github.com/aaa2ppp/contestio"
    "os"
)
func main() {
    br := NewReader(os.Stdin)
    bw := NewWriter(os.Stdout)
    defer bw.Flush()
    var n int
    ScanIntLn(br, &n)
    var a Ints[int]
    a = Resize(a, n)
    ScanSlice(br, a)
    PrintSliceLn(bw, a)
}`,
			buildTags:  "sugar",
			wantInline: true,
			wantClear:  true,
		},
		{
			name: "not_dot_import",
			content: `package main
import (
    "github.com/aaa2ppp/contestio"
    "os"
)
func main() {
    br := contestio.NewReader(os.Stdin)
    bw := contestio.NewWriter(os.Stdout)
    defer bw.Flush()
    var n int
    contestio.ScanIntLn(br, &n)
    var a contestio.Ints[int]
    a = contestio.Resize(a, n)
    contestio.ScanSlice(br, a)
    contestio.PrintSliceLn(bw, a)
}`,
			wantInline: false,
		},
		{
			name: "alias_import",
			content: `package main
import (
    cio "github.com/aaa2ppp/contestio"
    "os"
)
func main() {
    br := cio.NewReader(os.Stdin)
    bw := cio.NewWriter(os.Stdout)
    defer bw.Flush()
    var n int
    cio.ScanIntLn(br, &n)
    var a cio.Ints[int]
    a = cio.Resize(a, n)
    cio.ScanSlice(br, a)
    cio.PrintSliceLn(bw, a)
}`,
			wantInline: false,
		},
		{
			name: "no_import",
			content: `package main
import "os"
func main() {}`,
			wantInline: false,
		},
		{
			name: "already_inlined",
			content: `package main
import (
    . "github.com/aaa2ppp/contestio"
    "os"
)
// -- inline:github.com/aaa2ppp/contestio --
//...
// -- /inline:github.com/aaa2ppp/contestio --
func main() {
    br := NewReader(os.Stdin)
    bw := NewWriter(os.Stdout)
    defer bw.Flush()
    var n int
    ScanIntLn(br, &n)
    a := make([]int, n)
    ScanInts(br, a)
    PrintIntsLn(bw, a)
}`,
			wantInline: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			solDir := t.TempDir()
			mainPath := filepath.Join(solDir, "main.go")

			// Нормализуем и сохраняем исходник
			normalized, err := normalizeImports([]byte(tt.content), mainPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(mainPath, normalized, 0644); err != nil {
				t.Fatal(err)
			}

			// Настраиваем модуль
			mustRun(t, solDir, "go", "mod", "init", "example.com/solution")
			mustRun(t, solDir, "go", "mod", "edit", "-replace", "github.com/aaa2ppp/contestio="+libDir)
			mustRun(t, solDir, "go", "mod", "tidy")

			// Инлайн
			args := []string{mainPath}
			if tt.buildTags != "" {
				args = []string{"-tags=" + tt.buildTags, mainPath}
			}
			err = run(args)
			if (err == nil) != tt.wantInline {
				t.Fatalf("inline: got error %v, want success=%v", err, tt.wantInline)
			}
			if err != nil {
				return
			}

			// Проверки после инлайна
			checkAfterInline(t, mainPath, true)

			// Проверка компиляции
			if tt.buildTags != "" {
				mustRun(t, solDir, "go", "build", "-tags="+tt.buildTags, "-o", os.DevNull, ".")
			} else {
				mustRun(t, solDir, "go", "build", "-o", os.DevNull, ".")
			}

			if tt.skipClear {
				return
			}

			// Очистка
			clearArgs := []string{"-clear", mainPath}
			if tt.buildTags != "" {
				clearArgs = []string{"-clear", "-tags=" + tt.buildTags, mainPath}
			}
			err = run(clearArgs)
			if (err == nil) != tt.wantClear {
				t.Fatalf("clear: got error %v, want success=%v", err, tt.wantClear)
			}
			if err != nil {
				return
			}

			// Сравнение с оригиналом
			afterClear, err := os.ReadFile(mainPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(afterClear, normalized) {
				t.Errorf("after clear file differs from original")
			}
		})
	}
}

// checkAfterInline проверяет, что импорт удалён и вставлены маркеры.
func checkAfterInline(t *testing.T, mainPath string, wantImportRemoved bool) {
	t.Helper()
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("// -- inline:github.com/aaa2ppp/contestio --")) {
		t.Error("missing open inline marker")
	}
	if !bytes.Contains(data, []byte("// -- /inline:github.com/aaa2ppp/contestio --")) {
		t.Error("missing close inline marker")
	}
	if wantImportRemoved && bytes.Contains(data, []byte(`"github.com/aaa2ppp/contestio"`)) {
		t.Error("import still present")
	}
	// Проверка синтаксиса
	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, mainPath, data, parser.AllErrors)
	if err != nil {
		t.Errorf("syntax error: %v", err)
	}
}

// copyLibrary копирует .go файлы и go.mod из корня репозитория в dest.
func copyLibrary(dest string) error {
	root, err := filepath.Abs("../../")
	if err != nil {
		return err
	}
	files, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		dst := filepath.Join(dest, filepath.Base(f))
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return err
		}
	}
	// go.mod
	modSrc := filepath.Join(root, "go.mod")
	if data, err := os.ReadFile(modSrc); err == nil {
		dst := filepath.Join(dest, "go.mod")
		return os.WriteFile(dst, data, 0644)
	}
	return nil
}

// mustRun выполняет команду и требует успеха.
func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}
