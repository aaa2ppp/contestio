package main

import (
	"bufio"
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
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

const libPath = "github.com/aaa2ppp/contestio"

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [filename]\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  -clear    удалить встроенный код библиотеки\n")
	fmt.Fprintf(os.Stderr, "  -h        показать эту справку\n")
	fmt.Fprintf(os.Stderr, "\nfilename - файл, в который будет встроен код (по умолчанию main.go).\n")
	os.Exit(2)
}

func main() {
	log.SetFlags(0)

	var clear bool
	flag.BoolVar(&clear, "clear", false, "удалить встроенный код библиотеки")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	fileName := "main.go"
	if len(args) > 0 {
		fileName = args[0]
	}
	absFilePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatal(err)
	}

	if clear {
		if err := clearInliningFromFile(absFilePath, libPath); err != nil {
			log.Fatal(err)
		}
		log.Printf("Код библиотеки удалён из %s\n", fileName)
		return
	}

	// Этап 1: найти все экспортируемые объекты contestio, используемые в main.go
	pkg, err := loadMainPackage(absFilePath)
	if err != nil {
		log.Fatal(err)
	}
	rootNames := findRootObjectsInMain(pkg, libPath)
	if len(rootNames) == 0 {
		log.Fatalf("в файле %s не найдено обращений к экспортируемым объектам %s\n", fileName, libPath)
	}

	// Этап 2: загрузить пакет contestio, построить граф зависимостей и собрать все достижимые объекты
	pkg, err = loadLibPackage(libPath)
	if err != nil {
		log.Fatal(err)
	}
	nodeSet := extractDependencies(pkg, rootNames)
	if nodeSet == nil {
		log.Fatalf("ни один из корневых объектов в пакете %s не найден: %v\n", libPath, err)
	}

	// Этап 3: инлайним код найденных объектов в main.go
	if err := inlineNodeSetToFile(absFilePath, pkg, nodeSet); err != nil {
		log.Fatal(err)
	}
	log.Printf("Код библиотеки успешно встроен в %s\n", fileName)
}

func inlineNodeSetToFile(fileName string, pkg *packages.Package, nodeSet map[ast.Node]bool) error {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// создаем временный файл в том же каталоге
	tmpFile, err := os.CreateTemp(filepath.Dir(fileName), filepath.Base(fileName)+"-*")
	if err != nil {
		return err
	}
	defer func(fname string) {
		tmpFile.Close()
		os.Remove(fname)
	}(tmpFile.Name())

	if err := os.Chmod(tmpFile.Name(), fileInfo.Mode()); err != nil {
		return err
	}

	br := bufio.NewReader(file)
	bw := bufio.NewWriter(tmpFile)
	defer bw.Flush()

	// копируем файл исключая импрор пакета
	var isImport bool
	var foundImport bool
	for {
		s, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		// TODO: это хрупко
		t := strings.TrimSpace(s)
		if strings.HasPrefix(t, "import ") {
			isImport = true
		}
		if isImport && strings.Contains(s, pkg.PkgPath) {
			foundImport = true
		} else {
			bw.WriteString(s)
		}
		if isImport && (strings.HasSuffix(t, ")") ||
			strings.HasPrefix(t, "import ") && !strings.HasSuffix(t, "(")) {
			isImport = false
		}

		if err == io.EOF {
			break
		}
	}

	if !foundImport {
		return fmt.Errorf("not found import %s in file %s", pkg.PkgPath, fileName)
	}
	if err := file.Close(); err != nil {
		return err
	}

	// дописываем узлы пакета
	if err := printNodeSet(bw, pkg, nodeSet); err != nil {
		return fmt.Errorf("print nodes: %v", err)
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := fixImports(tmpFile.Name()); err != nil {
		return err
	}

	// делаем замену файла с бекапом
	if err := os.Rename(fileName, fileName+"~"); err != nil {
		return fmt.Errorf("can't backup %q file: %w", fileName, err)
	}
	if err := os.Rename(tmpFile.Name(), fileName); err != nil {
		os.Rename(fileName+"~", fileName)
		return fmt.Errorf("can't rename %q to %q", tmpFile.Name(), fileName)
	}

	return nil
}

func clearInliningFromFile(fileName string, pkgPath string) error {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	buf, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	var packagePos int
	if !bytes.HasPrefix(buf, []byte("package ")) {
		packagePos := bytes.Index(buf, []byte("\npackage "))
		if packagePos == -1 {
			return fmt.Errorf("not fond package in file %q", fileName)
		}
		packagePos++
	}

	importPos := bytes.Index(buf, []byte("\nimport "))
	if importPos != -1 && packagePos > importPos {
		return fmt.Errorf("not fond correct package/import in file %q", fileName)
	}
	importPos++

	openTag := "\n// -- inline:" + pkgPath + " --"
	openTagPos := bytes.Index(buf, []byte(openTag))
	if openTagPos == -1 || packagePos > openTagPos || importPos > openTagPos {
		return fmt.Errorf("not fond correct open tags %q position (%d) in file %q", openTag, openTagPos, fileName)
	}
	openTagPos++

	closeTag := "\n// -- /inline:" + pkgPath + " --"
	closeTagPos := bytes.Index(buf, []byte(closeTag))
	if closeTagPos == -1 || openTagPos > closeTagPos {
		return fmt.Errorf("not fond correct open/close tags in file %q", fileName)
	}
	closeTagPos++

	// создаем временный файл в том же каталоге
	tmpFile, err := os.CreateTemp(filepath.Dir(fileName), filepath.Base(fileName)+"-*")
	if err != nil {
		return err
	}
	defer func(fname string) {
		tmpFile.Close()
		os.Remove(fname)
	}(tmpFile.Name())

	if err := os.Chmod(tmpFile.Name(), fileInfo.Mode()); err != nil {
		return err
	}

	bw := bufio.NewWriter(tmpFile)
	defer bw.Flush()

	var end int

	if importPos == -1 {
		// нет ни одного импорта - всталяем отдельной строкой сразу за package
		end = bytes.IndexByte(buf[packagePos:], '\n') + packagePos + 1
		bw.Write(buf[:end])
		fmt.Fprintf(bw, "\nimport %q\n", pkgPath)
	} else {
		end = bytes.IndexByte(buf[importPos:], '\n') + importPos + 1
		bw.Write(buf[:end])

		t := bytes.TrimSpace(buf[importPos:end])
		if bytes.HasSuffix(t, []byte("(")) {
			// втавляем перевой строкой в блок импорта
			fmt.Fprintf(bw, "\t. %q\n", pkgPath)
		} else {
			// всталяем отдельной строкой
			fmt.Fprintf(bw, "\nimport . %q\n", pkgPath)
		}
	}

	// пропускаем все, что между строчками тегов, включая строчки тегов
	bw.Write(buf[end:openTagPos])
	end = bytes.IndexByte(buf[closeTagPos:], '\n') + closeTagPos + 1
	if end > closeTagPos {
		bw.Write(buf[end:])
	}

	if err := bw.Flush(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := fixImports(tmpFile.Name()); err != nil {
		return err
	}

	// делаем замену файла с бекапом
	if err := os.Rename(fileName, fileName+"~"); err != nil {
		return fmt.Errorf("can't backup %q file: %w", fileName, err)
	}
	if err := os.Rename(tmpFile.Name(), fileName); err != nil {
		os.Rename(fileName+"~", fileName)
		return fmt.Errorf("can't rename %q to %q", tmpFile.Name(), fileName)
	}

	return nil
}

func fixImports(fileName string) error {
	src, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	res, err := imports.Process(fileName, src, nil)
	if err != nil {
		return err
	}

	err = os.WriteFile(fileName, res, 0644)
	if err != nil {
		return err
	}

	return nil
}

// ---------------------------------------------------------------------
// Загрузка пакетов
// ---------------------------------------------------------------------

func loadMainPackage(filePath string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  filepath.Dir(filePath),
	}
	pkgs, err := packages.Load(cfg, "file="+filePath)
	if err != nil {
		return nil, fmt.Errorf("загрузка файла %s: %v", filePath, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("пакет %s не найден", filePath)
	}
	return pkgs[0], nil
}

func loadLibPackage(pkgPath string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("загрузка пакета %s: %v", pkgPath, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("пакет %s не найден", pkgPath)
	}
	return pkgs[0], err
}

// ---------------------------------------------------------------------
// Вывод собранных узлов
// ---------------------------------------------------------------------

func printPkgBoundary(w io.Writer, name string, open bool) error {
	var b bytes.Buffer
	b.Grow(82)
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

func printPkgOpenBoundary(w io.Writer, name string) error  { return printPkgBoundary(w, name, true) }
func printPkgCloseBoundary(w io.Writer, name string) error { return printPkgBoundary(w, name, false) }

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

	printPkgOpenBoundary(w, pkg.PkgPath)

	var currentFile string
	for _, it := range nodes {
		baseName := filepath.Base(it.pos.Filename)
		if baseName != currentFile {
			currentFile = baseName
			fmt.Fprintf(w, "\n// == %s ==\n\n", currentFile)
		}
		var buf bytes.Buffer
		if err := format.Node(&buf, pkg.Fset, it.node); err != nil {
			fmt.Fprintf(os.Stderr, "ошибка форматирования: %v\n", err)
			continue
		}
		fmt.Fprintln(w, buf.String())
		fmt.Fprintln(w) // пустая строка между объявлениями
	}

	return printPkgCloseBoundary(w, pkg.PkgPath)
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
