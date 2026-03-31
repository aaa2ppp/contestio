BIN_DIR ?= ./bin
TMP_DIR ?= ./tmp
BENCH_DIR ?= ./benchmarks

GOEXE := $(shell go env GOEXE)

# build tags
TAGS ?= dev,sugar,any

MERGE_FILES ?= Makefile go.mod go.sum *.go *.sh *.md *.txt

# source for merge/patch operations
SRC ?= .
# destination for merge/patch operations
DST ?= 1

# Кастомные флаги сборки (можно переопределить при вызове make)
BUILD_FLAGS ?=

# Находим все поддиректории в cmd, которые потенциально могут быть бинарниками
CMDS := $(wildcard cmd/*)

# Генерируем список целей для бинарников
BINARIES := $(patsubst cmd/%,$(BIN_DIR)/%,$(CMDS))


.PHONY: FORCE all deps test benchgit  build clean run

# Основная цель - собирает все бинарники
all: deps test build ## update deps, test and build all

FORCE:

# Правило для подготовки зависимостей
deps: ## update deps
	go mod tidy

.PHONY: test-contestio-inline test-lib-unsafe test-lib

test-inline: ## do contestio-inline tests
	go test -tags=$(TAGS) ./cmd/contestio-inline

test-lib: ## do lib tests without `unsafe` tag
	go test -tags=$(TAGS) .

test-lib-unsafe: ## do lib tests with `unsafe` tag
	go test -tags=$(TAGS),unsafe .

test: test-lib test-lib-unsafe ## do all lib tests
	@echo OK

GO_BENCH = go test -tags=$(TAGS) -run '^$$' -benchmem -bench

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
	$(GO_BENCH) 'parseInt$(subst bench-parse-int,,$@)$$' .

# Функции капитализации (int -> Int, float -> Float)
capitalize_int   = Int
capitalize_float = Float


# Статические правила для scan
$(SCAN_TARGETS): bench-scan-%:
	$(GO_BENCH) 'scan$(capitalize_$*)$$' .

# Статические правила для print
$(PRINT_TARGETS): bench-print-%:
	$(GO_BENCH) 'print$(capitalize_$*)$$' .

# Групповые цели
bench-parse-int-all: $(PARSE_TARGETS) ## do all parse benchmarks
bench-scan-all: $(SCAN_TARGETS)       ## do all scan benchmarks
bench-print-all: $(PRINT_TARGETS)     ## do all print benchmarks

benchmarks: FORCE ## create benchmark reports
	@mkdir -p $(BENCH_DIR)
	@$(MAKE) bench-parse-int-all > $(BENCH_DIR)/parse-int-all
	@$(MAKE) bench-scan-all      > $(BENCH_DIR)/scan-all
	@$(MAKE) bench-print-all     > $(BENCH_DIR)/print-all

# Шаблонное правило для сборки любого бинарника
$(BIN_DIR)/%: FORCE
	@mkdir -p $(@D)
	go build $(BUILD_FLAGS) -o $@$(GOEXE) ./cmd/$(notdir $@)

build: $(BINARIES) ## build all binaries

clean: ## remove temporary and binary files
	-rm -rf $(BIN_DIR) $(TMP_DIR) $(BENCH_DIR)


.PHONY: merge patch

MERGE_FIND_PARTS := $(patsubst %,-o -name '%',$(MERGE_FILES))
MERGE_FIND_EXPR := $(wordlist 2,$(words $(MERGE_FIND_PARTS)),$(MERGE_FIND_PARTS))

merge: ## merge code to one file
	@mkdir -p $(TMP_DIR)
	@find $(SRC) -type f ! -path '*/qmail-src/src/*' \( $(MERGE_FIND_EXPR) \) -exec sh -c 'name="{}"; printf "== $${name#./} ==\n\n"; cat $$name; echo' ';' > $(TMP_DIR)/$(DST).code
	@echo "Merge saved to $(TMP_DIR)/$(DST).code"	
	

patch: bump-note-id deps test ## make precommit patch
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


.PHONY: bump-note-id notes notes-sort
bump-note-id: ## update next note id
	@next=$$(grep -Eo '[0-9]{3,} \[.?\]' NOTES.md | sort -n | tail -1 | awk '{printf "%03d", $$1+1}'); \
	sed -i "s/<!-- next-note-id:[0-9]\{3,\} -->/<!-- next-note-id:$$next -->/" NOTES.md

notes: ## show notes
	@grep -E '^(- \*\*[0-9]{3}|##|<!-- next-note-id:)' NOTES.md

notes-sort: ## show sorted notes
	@grep -E '^\- \*\*[0-9]{3}' NOTES.md | sort && \
	grep '<!-- next-note-id:' NOTES.md

help: ## show this help
	@printf "Usage: make [target] [VARIABLE=value]\n\n"
	@printf "Variables:\n"
	@awk 'BEGIN {comment=""} \
		/^[a-zA-Z0-9_-]+[[:space:]]*\?=/ { \
			split($$0, a, "?="); \
			if ( prev ~ /^#/ ) { \
				printf "  %-14s = %-20s %s\n", a[1], a[2], prev; \
			} else { \
				printf "  %-14s = %-20s\n", a[1], a[2]; \
			} \
		} \
		{ prev=$$0 }' $(MAKEFILE_LIST)
	@printf "\nTargets:\n"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  %-20s - %s\n", $$1, $$2}' $(MAKEFILE_LIST)
