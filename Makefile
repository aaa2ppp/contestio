BIN_DIR ?= ./bin
TMP_DIR ?= ./tmp
BENCH_DIR ?= ./benchmarks

GOEXE := $(shell go env GOEXE)
TEST_FLAGS ?= -tags=dev,sugar
BENCH_FLAGS ?= -benchmem 

MERGE_FILES ?= Makefile go.mod go.sum *.go *.sh *.md *.txt

# source and destination for merge/patch operations
SRC ?= .
DST ?= 1

# Кастомные флаги сборки (можно переопределить при вызове make)
BUILD_FLAGS ?=

# Находим все поддиректории в cmd, которые потенциально могут быть бинарниками
CMDS := $(wildcard cmd/*)

# Генерируем список целей для бинарников
BINARIES := $(patsubst cmd/%,$(BIN_DIR)/%,$(CMDS))


.PHONY: FORCE all deps test benchgit  build clean run

# Основная цель - собирает все бинарники
all: deps test build

FORCE:

# Правило для подготовки зависимостей
deps:
	go mod tidy

test:
	go test $(TEST_FLAGS) ./...
	@echo OK

# Суффиксы для parseInt (пустой для базовой цели)
PARSE_SUFFIXES := 2 4 6 8 12 16
PARSE_TARGETS  := bench-parse-int $(addprefix bench-parse-int, $(PARSE_SUFFIXES))

# Общие суффиксы для scan и print
SCAN_PRINT_SUFFIXES := int float
SCAN_TARGETS  := $(addprefix bench-scan-, $(SCAN_PRINT_SUFFIXES))
PRINT_TARGETS := $(addprefix bench-print-, $(SCAN_PRINT_SUFFIXES))

# Все цели бенчмарков
ALL_BENCH_TARGETS := $(PARSE_TARGETS) $(SCAN_TARGETS) $(PRINT_TARGETS)

# Явно объявляем все цели .PHONY
.PHONY: $(ALL_BENCH_TARGETS) bench-parse-int-all bench-scan-all bench-print-all

# Статическое правило для parseInt
$(PARSE_TARGETS): %:
	go test -bench 'parseInt$(subst bench-parse-int,,$@)$$' $(BENCH_FLAGS)

# Функции капитализации (int -> Int, float -> Float)
capitalize_int   = Int
capitalize_float = Float

# Статические правила для scan
$(SCAN_TARGETS): bench-scan-%:
	go test -bench 'scan$(capitalize_$*)$$' $(BENCH_FLAGS)

# Статические правила для print
$(PRINT_TARGETS): bench-print-%:
	go test -bench 'print$(capitalize_$*)$$' $(BENCH_FLAGS)

# Групповые цели
bench-parse-int-all: $(PARSE_TARGETS)
bench-scan-all: $(SCAN_TARGETS)
bench-print-all: $(PRINT_TARGETS)

benchmarks: FORCE
	@mkdir -p $(BENCH_DIR)
	@$(MAKE) bench-parse-int-all > $(BENCH_DIR)/parse-int-all
	@$(MAKE) bench-scan-all      > $(BENCH_DIR)/scan-all
	@$(MAKE) bench-print-all     > $(BENCH_DIR)/print-all

# Шаблонное правило для сборки любого бинарника
$(BIN_DIR)/%: FORCE
	@mkdir -p $(@D)
	go build $(BUILD_FLAGS) -o $@$(GOEXE) ./cmd/$(notdir $@)

build: $(BINARIES)

# Очистка
clean:
	-rm -rf $(BIN_DIR) $(TMP_DIR) $(BENCH_DIR)


.PHONY: merge patch

MERGE_FIND_PARTS := $(patsubst %,-o -name '%',$(MERGE_FILES))
MERGE_FIND_EXPR := $(wordlist 2,$(words $(MERGE_FIND_PARTS)),$(MERGE_FIND_PARTS))

merge:
	@mkdir -p $(TMP_DIR)
	@find $(SRC) -type f ! -path '*/qmail-src/src/*' \( $(MERGE_FIND_EXPR) \) -exec sh -c 'name="{}"; printf "== $${name#./} ==\n\n"; cat $$name; echo' ';' > $(TMP_DIR)/$(DST).code
	@echo "Merge saved to $(TMP_DIR)/$(DST).code"	
	

# Создает прекоммит патч
patch: bump-note-id deps test
	@mkdir -p $(TMP_DIR)
	
	@(set -e; \
	staged_list="$(TMP_DIR)/staged_list.$$$$"; \
	unstaged_list="$(TMP_DIR)/unstaged_list.$$$$"; \
	git diff --staged --name-only -- $(SRC) > "$$staged_list"; \
	git diff --name-only -- $(SRC) > "$$unstaged_list"; \
	intersection=$$(grep -Fxf "$$staged_list" "$$unstaged_list" || true); \
	rm -f "$$staged_list" "$$unstaged_list"; \
	if [ -n "$$intersection" ]; then \
		echo "" >&2; \
		echo "WARNING: the following files have changes not staged for commit:" >&2; \
		echo "  (use \"git add <file>...\" to update what will be committed)" >&2; \
		printf '%s\n' $$intersection | sed 's/^/        /' >&2; \
		echo "" >&2; \
	fi)
	
	git diff --staged -- $(SRC) > $(TMP_DIR)/$(DST).patch
	@echo "Patch saved to $(TMP_DIR)/$(DST).patch"


.PHONY: bump-note-id
bump-note-id:
	@next=$$(grep -Eo '[0-9]{3,} \[.?\]' NOTES.md | sort -n | tail -1 | awk '{printf "%03d", $$1+1}'); \
	sed -i "s/<!-- next-note-id:[0-9]\{3,\} -->/<!-- next-note-id:$$next -->/" NOTES.md
