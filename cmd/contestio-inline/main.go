package main

import (
	"bytes"
	"cmp"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

const libPath = "github.com/aaa2ppp/contestio"

func usage(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintf(fs.Output(), "Usage: %s [options] [filename]\n", fs.Name())
		fmt.Fprintf(fs.Output(), "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nfilename - файл, в который будет встроен код (по умолчанию main.go).\n")
		os.Exit(2)
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	log.SetFlags(0)

	fs := flag.NewFlagSet("contestio-inline", flag.ContinueOnError)
	var clear bool
	var buildTags string
	var noBuildCheck bool

	fs.BoolVar(&clear, "clear", false, "удалить встроенный код библиотеки")
	fs.BoolVar(&noBuildCheck, "no-build-check", false, "отключить проверку компиляции")
	fs.StringVar(&buildTags, "tags", "", "теги сборки (см. go help build)")
	fs.Usage = usage(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}

	opts := inlineOpts{
		buildTags:    buildTags,
		noBuildCheck: noBuildCheck,
	}

	posArgs := fs.Args()
	fileName := "main.go"
	if len(posArgs) > 0 {
		fileName = posArgs[0]
		info, err := os.Stat(fileName)
		if err != nil {
			return err
		}
		if info.IsDir() {
			fileName = filepath.Join(fileName, "main.go")
		}
	}
	absFilePath, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	solutionDir := filepath.Dir(absFilePath)

	if clear {
		if err := clearInliningFromFile(absFilePath, libPath, opts); err != nil {
			return err
		}
		log.Printf("Код библиотеки удалён из %s\n", fileName)
		return nil
	}

	// Этап 1: найти все экспортируемые объекты contestio, используемые в main.go
	pkg, err := loadPackage("file="+absFilePath, buildTags, solutionDir)
	if err != nil {
		return err
	}
	rootNames := findRootObjectsInMain(pkg, libPath)
	if len(rootNames) == 0 {
		return fmt.Errorf("В файле %s не найдено обращений к объектам пакета %s", fileName, libPath)
	}

	// Этап 2: загрузить пакет contestio, построить граф зависимостей и собрать все достижимые объекты
	pkg, err = loadPackage(libPath, buildTags, solutionDir)
	if err != nil {
		return err
	}
	nodeSet := extractDependencies(pkg, rootNames)
	if nodeSet == nil {
		return fmt.Errorf("В пакете %s не найдено ни одиного корневого объекта: %v", libPath, err)
	}

	// Этап 3: инлайним код найденных объектов в main.go
	if err := inlineNodeSetToFile(absFilePath, pkg, nodeSet, opts); err != nil {
		return err
	}
	log.Printf("Код библиотеки успешно встроен в %s\n", fileName)

	return nil
}

type inlineOpts struct {
	buildTags    string
	noBuildCheck bool
}

func inlineNodeSetToFile(fileName string, pkg *packages.Package, nodeSet map[ast.Node]bool, opts inlineOpts) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	input, err = normalizeImports(input, fileName)
	if err != nil {
		return fmt.Errorf("normalize imports: %v", err)
	}

	openTagPos, closeTagPos := findInlineTags(input, pkg.PkgPath)
	if openTagPos != -1 || closeTagPos != -1 {
		return fmt.Errorf("File %q already contains open or close inline tag", fileName)
	}

	var outbuf bytes.Buffer

	// копируем файл исключая импорт пакета
	var isImport bool
	var foundImport bool
	for pos := 0; pos < len(input); {
		end := bytes.IndexByte(input[pos:], '\n') + pos + 1
		if end == pos {
			end = len(input)
		}
		s := string(input[pos:end])
		pos = end

		// TODO: это хрупко
		target := ". " + strconv.Quote(pkg.PkgPath)
		t := strings.TrimSpace(s)
		if strings.HasPrefix(t, "import ") {
			isImport = true
		}
		if isImport && strings.Contains(s, target) {
			foundImport = true
		} else {
			outbuf.WriteString(s)
		}
		if isImport && (strings.HasSuffix(t, ")") ||
			strings.HasPrefix(t, "import ") && !strings.HasSuffix(t, "(")) {
			isImport = false
		}
	}

	if !foundImport {
		return fmt.Errorf("Not found dot import %q in file %q", pkg.PkgPath, fileName)
	}

	// дописываем узлы пакета
	if err := printNodeSet(&outbuf, pkg, nodeSet); err != nil {
		return fmt.Errorf("print nodes: %v", err)
	}

	return rewriteFileWithBackup(fileName, outbuf.Bytes(), opts)
}

// compileCheck проверяет, компилируется ли переданный код Go.
func compileCheck(code []byte, buildTags string) error {
	tmpFile, err := os.CreateTemp("", "check-*.go")
	if err != nil {
		return err
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(code); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	nullDevice := os.DevNull
	cmd := exec.Command("go", "build", "-tags="+buildTags, "-o", nullDevice, tmpFile.Name())

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("FAIL: build -tags=%s\n%s", buildTags, stderr.String())
	}
	return nil
}

func findLinePrefix(buf []byte, prefix string) int {
	if bytes.HasPrefix(buf, []byte(prefix)) {
		return 0
	}
	pos := bytes.Index(buf, []byte("\n"+prefix))
	if pos == -1 {
		return -1
	}
	return pos + 1
}

func findInlineTags(buf []byte, pkgPath string) (int, int) {
	openPos := findLinePrefix(buf, "// -- inline:"+pkgPath+" --")
	closePos := findLinePrefix(buf, "// -- /inline:"+pkgPath+" --")
	return openPos, closePos
}

func clearInliningFromFile(fileName string, pkgPath string, opts inlineOpts) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	packagePos := findLinePrefix(input, "package ")
	if packagePos == -1 {
		return fmt.Errorf("Not found package in file %q", fileName)
	}

	importPos := findLinePrefix(input, "import ")
	if importPos != -1 && packagePos > importPos {
		return fmt.Errorf("Not found correct import in file %q", fileName)
	}

	openTagPos, closeTagPos := findInlineTags(input, pkgPath)
	if packagePos > openTagPos || importPos > openTagPos || openTagPos > closeTagPos {
		return fmt.Errorf("Not found correct open and close inline tags in file %q", fileName)
	}

	var outbuf bytes.Buffer
	var end int

	if importPos == -1 {
		// нет ни одного импорта - вставляем отдельной строкой сразу за package
		end = bytes.IndexByte(input[packagePos:], '\n') + packagePos + 1
		outbuf.Write(input[:end])
		fmt.Fprintf(&outbuf, "\nimport %q\n", pkgPath)
	} else {
		end = bytes.IndexByte(input[importPos:], '\n') + importPos + 1
		outbuf.Write(input[:end])

		t := bytes.TrimSpace(input[importPos:end])
		if bytes.HasSuffix(t, []byte("(")) {
			// вставляем перевой строкой в блок импорта
			fmt.Fprintf(&outbuf, "\t. %q\n", pkgPath)
		} else {
			// вставляем отдельной строкой
			fmt.Fprintf(&outbuf, "\nimport . %q\n", pkgPath)
		}
	}

	// пропускаем все, что между строчками тегов, включая строчки тегов
	outbuf.Write(input[end:openTagPos])
	end = bytes.IndexByte(input[closeTagPos:], '\n') + closeTagPos + 1
	if end > closeTagPos {
		outbuf.Write(input[end:])
	}

	return rewriteFileWithBackup(fileName, outbuf.Bytes(), opts)
}

func rewriteFileWithBackup(fileName string, output []byte, opts inlineOpts) error {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	output, err = normalizeImports(output, fileName)
	if err != nil {
		return fmt.Errorf("final normalize imports: %v", err)
	}

	if !opts.noBuildCheck {
		if err := compileCheck(output, opts.buildTags); err != nil {
			return err
		}
	}

	if err := os.Rename(fileName, fileName+"~"); err != nil {
		return fmt.Errorf("Can't backup %q file: %w", fileName, err)
	}

	if err := os.WriteFile(fileName, output, fileInfo.Mode()); err != nil {
		os.Rename(fileName+"~", fileName)
		return err
	}

	return nil
}

func normalizeImports(src []byte, filename string) ([]byte, error) {
	return imports.Process(filename, src, nil)
}

// ---------------------------------------------------------------------
// Загрузка пакетов
// ---------------------------------------------------------------------

// loadPackage загружает пакет по заданному шаблону (например, "file=main.go" или "github.com/aaa2ppp/contestio")
// в контексте директории dir с указанными тегами сборки.
func loadPackage(pattern, buildTags, dir string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:        dir,
		BuildFlags: []string{"-tags=" + buildTags},
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("загрузка %s: %v", pattern, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("пакет %s не найден", pattern)
	}
	return pkgs[0], nil
}

// ---------------------------------------------------------------------
// Вывод собранных узлов
// ---------------------------------------------------------------------

func printInlineTag(w io.Writer, name string, open bool) error {
	var b bytes.Buffer
	b.Grow(83)
	if open {
		b.WriteString("\n// -- inline:")
	} else {
		b.WriteString("\n// -- /inline:")
	}
	b.WriteString(name)
	b.WriteString(" --")
	for b.Len() < 81 {
		b.WriteByte('-')
	}
	b.WriteByte('\n')
	_, err := b.WriteTo(w)
	return err
}

func printOpenInlineTag(w io.Writer, name string) error  { return printInlineTag(w, name, true) }
func printCloseInlineTag(w io.Writer, name string) error { return printInlineTag(w, name, false) }

func printNodeSet(w io.Writer, pkg *packages.Package, nodeSet map[ast.Node]bool) error {
	// sort node positions
	type nodePosition struct {
		node ast.Node
		pos  token.Position
	}
	nodes := make([]nodePosition, 0, len(nodeSet))
	for node := range nodeSet {
		switch node.(type) {
		case *ast.GenDecl, *ast.FuncDecl:
			nodePos := nodePosition{node, pkg.Fset.Position(node.Pos())}
			nodes = append(nodes, nodePos)
		default:
			log.Printf("printNodeSet: skip node %[1]v is %[1]T, want *ast.GenDecl or *ast.FuncDecl", node)
		}
	}
	slices.SortFunc(nodes, func(a, b nodePosition) int {
		return cmp.Or(strings.Compare(a.pos.Filename, b.pos.Filename), a.pos.Offset-b.pos.Offset)
	})

	printOpenInlineTag(w, pkg.PkgPath)
	fmt.Fprintln(w)

	var buf bytes.Buffer
	for _, it := range nodes {
		removeComments(it.node)
		buf.Reset()
		if err := format.Node(&buf, pkg.Fset, it.node); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка форматирования: %v\n", err)
			continue
		}
		w.Write(removeBlankLines(buf.Bytes()))
	}

	return printCloseInlineTag(w, pkg.PkgPath)
}

// removeBlankLines works inplace uses the space of the same slice for writing
func removeBlankLines(b []byte) []byte {
	lines := bytes.Split(b, []byte("\n"))
	b = b[:0]
	for _, l := range lines {
		if len(bytes.TrimRight(l, " \t\r\n")) == 0 {
			continue
		}
		b = append(b, l...)
		b = append(b, '\n')
	}
	return b
}

// removeComments рекурсивно удаляет все комментарии из узла AST.
func removeComments(node ast.Node) {
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		// Обнуляем комментарии у всех возможных полей
		switch x := n.(type) {
		case *ast.File:
			x.Doc = nil
			x.Comments = nil
		case *ast.GenDecl:
			x.Doc = nil
		case *ast.FuncDecl:
			x.Doc = nil
		case *ast.TypeSpec:
			x.Doc = nil
			x.Comment = nil
		case *ast.ValueSpec:
			x.Doc = nil
			x.Comment = nil
		case *ast.Field:
			x.Doc = nil
			x.Comment = nil
		case *ast.FieldList:
			// поля списка обрабатываются отдельно, но у самого FieldList комментариев нет
		case *ast.BlockStmt:
			// у блока могут быть комментарии? в go/ast нет поля Comment у BlockStmt
		case *ast.CommentGroup:
			// такие узлы мы вообще не хотим видеть, но если встретились, можно игнорировать
		}
		return true
	})
}

// ---------------------------------------------------------------------
// Этап 1: поиск корневых объектов в main.go
// ---------------------------------------------------------------------

// findRootObjectsInMain возвращает имена экспортируемых объектов из пакета библиотеки,
// которые встречаются в коде (идентификаторы, связанные с объектами целевого пакета).
func findRootObjectsInMain(pkg *packages.Package, libPkgPath string) map[string]bool {
	// Находим среди импортированных пакетов contestio
	var targetPkg *types.Package
	for _, imp := range pkg.Types.Imports() {
		if imp.Path() == libPkgPath {
			targetPkg = imp
			break
		}
	}
	if targetPkg == nil {
		return nil
	}

	// Обходим AST в поисках идентификаторов, ссылающихся на экспортируемые объекты contestio
	names := make(map[string]bool)
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			ident, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pkg.TypesInfo.ObjectOf(ident)
			if obj == nil || !obj.Exported() || obj.Pkg() != targetPkg {
				return true
			}
			names[obj.Name()] = true
			return true
		})
	}
	return names
}

// ---------------------------------------------------------------------
// Этап 2: извлечение зависимостей из пакета contestio
// ---------------------------------------------------------------------

// extractDependencies строит граф зависимостей между его объектами,
// находит все объекты, достижимые из корневых (rootNames), и возвращает их.
func extractDependencies(pkg *packages.Package, rootNames map[string]bool) map[ast.Node]bool {
	// -------------------------------------------------------------
	// Шаг 1: построение графа зависимостей и карты объявлений
	// -------------------------------------------------------------
	deps, declOf := buildGraph(pkg)

	// -------------------------------------------------------------
	// Шаг 2: поиск корневых объектов (экспортированные имена из rootNames)
	// -------------------------------------------------------------
	scope := pkg.Types.Scope()
	var roots []types.Object
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj.Exported() && rootNames[obj.Name()] {
			roots = append(roots, obj)
		}
	}
	if len(roots) == 0 {
		return nil
	}

	// -------------------------------------------------------------
	// Шаг 3: обход графа в ширину от корней
	// -------------------------------------------------------------
	visited := bfs(roots, deps)

	// -------------------------------------------------------------
	// Шаг 4: принудительно добавляем все методы для типов, попавших в visited
	// -------------------------------------------------------------
	addAllMethods(visited, pkg.Types)

	// -------------------------------------------------------------
	// Шаг 5: фильтрация – оставляем только глобальные объекты (принадлежащие пакету или методы)
	// -------------------------------------------------------------
	globalVisited := filterGlobal(visited, scope)

	// -------------------------------------------------------------
	// Шаг 6: сбор уникальных узлов AST для вывода
	// -------------------------------------------------------------
	declNodeSet := collectDeclarations(globalVisited, declOf)

	return declNodeSet
}

// ---------------------------------------------------------------------
// Вспомогательные функции для построения графа и обхода
// ---------------------------------------------------------------------

// buildGraph обходит AST пакета, собирает информацию:
//   - deps[obj] — множество объектов, от которых зависит obj (через использования идентификаторов)
//   - declOf[obj] — узел объявления верхнего уровня (FuncDecl или GenDecl), где определён obj
//
// При обходе отслеживается текущий объект-владелец (owner), чтобы правильно связывать использования.
func buildGraph(pkg *packages.Package) (deps map[types.Object]map[types.Object]bool, declOf map[types.Object]ast.Node) {
	deps = make(map[types.Object]map[types.Object]bool)
	declOf = make(map[types.Object]ast.Node)

	// Для каждого файла в пакете
	for _, file := range pkg.Syntax {
		// Рекурсивный обход с параметром owner (текущий определяемый объект)
		var visit func(n ast.Node, owner types.Object)
		visit = func(n ast.Node, owner types.Object) {
			if n == nil {
				return
			}

			switch node := n.(type) {
			// Обработка объявления функции или метода
			case *ast.FuncDecl:
				obj := pkg.TypesInfo.Defs[node.Name]
				if obj != nil && obj.Pkg() == pkg.Types {
					declOf[obj] = node // сохраняем узел объявления
					// Обрабатываем ресивер, параметры и тело с новым владельцем obj
					if node.Recv != nil {
						visit(node.Recv, obj)
					}
					if node.Type != nil {
						visit(node.Type, obj)
					}
					if node.Body != nil {
						visit(node.Body, obj)
					}
				}
				return // избегаем повторного обхода через общий код

			// Обработка общих объявлений (type, var, const)
			case *ast.GenDecl:
				// Определяем, является ли объявление глобальным (owner == nil)
				isGlobal := owner == nil
				for _, spec := range node.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						obj := pkg.TypesInfo.Defs[spec.Name]
						if obj != nil && obj.Pkg() == pkg.Types {
							if isGlobal {
								declOf[obj] = node // сохраняем для глобального типа
							}
							if spec.Type != nil {
								visit(spec.Type, obj) // обрабатываем тип с владельцем obj
							}
						}
					case *ast.ValueSpec:
						// В одной спецификации может быть несколько имён (var x, y int)
						for _, name := range spec.Names {
							obj := pkg.TypesInfo.Defs[name]
							if obj != nil && obj.Pkg() == pkg.Types {
								if isGlobal {
									declOf[obj] = node // сохраняем для глобальной переменной/константы
								}
								// Обрабатываем тип и значения для каждого имени отдельно,
								// чтобы зависимости привязывались к конкретному объекту
								if spec.Type != nil {
									visit(spec.Type, obj)
								}
								for _, val := range spec.Values {
									if val != nil {
										visit(val, obj)
									}
								}
							}
						}
					}
				}
				return

			// Обработка идентификатора – регистрируем использование
			case *ast.Ident:
				if obj := pkg.TypesInfo.Uses[node]; obj != nil && obj.Pkg() == pkg.Types {
					if owner != nil {
						if deps[owner] == nil {
							deps[owner] = make(map[types.Object]bool)
						}
						deps[owner][obj] = true
					}
				}
			}

			// Рекурсивно обходим дочерние узлы с тем же владельцем
			ast.Inspect(n, func(child ast.Node) bool {
				if child != nil && child != n {
					visit(child, owner)
				}
				return true
			})
		}

		// Начинаем обход файла с владельцем nil
		visit(file, nil)
	}

	// После обработки AST добавляем явные зависимости от типа к его методам.
	// Это нужно, потому что методы не всегда явно используются в коде,
	// но они считаются частью типа и должны быть включены.
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if typeName, ok := obj.(*types.TypeName); ok {
			if named, ok := typeName.Type().(*types.Named); ok {
				for i := 0; i < named.NumMethods(); i++ {
					method := named.Method(i)
					if method.Pkg() == pkg.Types {
						if deps[typeName] == nil {
							deps[typeName] = make(map[types.Object]bool)
						}
						deps[typeName][method] = true
					}
				}
			}
		}
	}

	return deps, declOf
}

// bfs выполняет обход графа в ширину от стартовых объектов и возвращает множество достижимых.
func bfs(roots []types.Object, deps map[types.Object]map[types.Object]bool) map[types.Object]bool {
	visited := make(map[types.Object]bool)
	queue := append([]types.Object{}, roots...)
	for len(queue) > 0 {
		obj := queue[0]
		queue = queue[1:]
		if visited[obj] {
			continue
		}
		visited[obj] = true
		for dep := range deps[obj] {
			if !visited[dep] {
				queue = append(queue, dep)
			}
		}
	}
	return visited
}

// addAllMethods добавляет в visited все методы для каждого типа, уже присутствующего в visited.
// Это гарантирует, что даже если метод не использовался явно, он будет включён.
func addAllMethods(visited map[types.Object]bool, pkg *types.Package) {
	// Сначала соберём все типы из visited
	typesSet := make(map[*types.TypeName]bool)
	for obj := range visited {
		if typeName, ok := obj.(*types.TypeName); ok {
			typesSet[typeName] = true
		}
	}
	// Для каждого типа добавим все его методы
	for typeName := range typesSet {
		if named, ok := typeName.Type().(*types.Named); ok {
			for i := 0; i < named.NumMethods(); i++ {
				method := named.Method(i)
				if method.Pkg() == pkg {
					visited[method] = true
				}
			}
		}
	}
}

// filterGlobal оставляет только объекты, которые являются глобальными в пакете:
//   - методы (функции с получателем) включаются всегда;
//   - все остальные объекты включаются только если их областью видимости является пакетный scope.
//
// Это отсеивает локальные переменные, константы и типы, объявленные внутри функций.
func filterGlobal(visited map[types.Object]bool, pkgScope *types.Scope) map[types.Object]bool {
	global := make(map[types.Object]bool)
	for obj := range visited {
		// Проверяем, является ли объект методом
		if isMethod(obj) {
			global[obj] = true
			continue
		}
		// Для всех остальных – проверяем принадлежность пакетному scope
		if obj.Parent() == pkgScope {
			global[obj] = true
		}
	}
	return global
}

// isMethod возвращает true, если объект – функция и у неё есть получатель.
func isMethod(obj types.Object) bool {
	funcObj, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	// У методов сигнатура содержит получатель
	sig := funcObj.Type().(*types.Signature)
	return sig.Recv() != nil
}

// collectDeclarations собирает уникальные узлы AST, соответствующие объявлениям объектов.
func collectDeclarations(objects map[types.Object]bool, declOf map[types.Object]ast.Node) map[ast.Node]bool {
	nodes := make(map[ast.Node]bool)
	for obj := range objects {
		if decl := declOf[obj]; decl != nil {
			nodes[decl] = true
		}
	}
	return nodes
}
