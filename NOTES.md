# Notes on the "Contest IO" Project

<!-- next-note-id:022 -->

## Open Questions

- **004 [ ] Public SkipSpace/SkipSpaceLn/SkipLn (2026-03-09)**

  Should we make these functions public?
  Currently, this functionality is implemented via private functions `skipSpace`/`skipToNewLine` (used inside `scanXXX`).  
  Users may want to explicitly use them in their code, but I haven't encountered any such cases yet.

- **005 [ ] Move library code out of root? (2026-03-09)**

  Currently all library files are in the root. Should we move them to a subdirectory to keep root clean?  
  If so, what name? I dislike plain `lib`. Perhaps just keep as is and move examples to `examples/`?

- **006 [ ] What to do with Sugar? (2026-03-10)**

  The `ScanSlice` and `PrintSlice` functions accept `Parser` and `ValAppender` interfaces, respectively.
  This allows scanning a slice of any type using only these functions.
  The solution seems heavyweight (performance overhead) and clunky (requires implementing an interface).
  Currently, this functionality is isolated from the main codebase with the build tag `-tags=sugar`.
  Should we delete it or merge it into the main codebase?

- **017 [ ] Add streaming iterators? (2026-03-21)**

  Currently we have `ScanInts`, `ScanFloats`, `ScanWords` that fill or create slices. For streaming scenarios where storing the whole slice is unnecessary, iterators would allow processing elements one by one without allocating the slice.

  **Key questions:**
  - Are iterators needed at all, or does the slice API already cover most use cases?
  - For `Word`, we could return a string backed by the internal buffer (zero‑copy, no allocation), but this ties the string’s lifetime to the buffer. Is this safety trade‑off acceptable? Alternatively, copying per token would allocate, defeating the streaming benefit.
  - Should iterators be part of the main library or remain an experimental feature?

  **API style** must be Go 1.23 iterators (`iter.Seq[T]`) to enable direct use with `range`.  

- **008 [ ] Should nextToken/scanXxx advance position on read/parse errors? (2026-03-11)**

  Currently, `nextToken` always discards the bytes it has consumed, even if it returns an error (e.g., `ErrTokenTooLong` or `io.EOF` after a partial token). Similarly, `scanSlice`/`scanVars` advance the reader past the token that caused a parse error before returning the error. This behavior is undocumented.

  **Pros of current approach:**
  - Simple and consistent: after an error, the stream is left at the next possible position.
  - Useful for resuming after a recoverable error (e.g., skipping malformed input).

  **Cons:**
  - If the caller wants to inspect or retry the erroneous token, the data is lost.
  - Error recovery becomes non-trivial because the state is already mutated.

  Should we change the behavior to **not advance** on parse errors (i.e., leave the reader pointing at the beginning of the offending token)? Or should we keep the current behavior but document it clearly? This decision affects all scanning functions and the `nextToken` helper.


  
## Ideas

## Plans

- **001 [ ] Refactor `contestio-inline` (2026-03-09)**

  Complex auto-generated parts (e.g. AST traversal) to be clear and maintainable: simplify structure, cover key scenarios with tests.

- **002 [ ] Rework the README (2026-03-09)**

  Important: 
    - **what it is** – briefly
    - **key features** – briefly
    - **quick start** – detailed  
    - literary flourishes and reflections to the footer

  English version of the README.

- **003 [ ] Translate all inline documentation to English (2026-03-09)**

- **021 [ ] Investigate unification of scan and output cycles (2026-03-28)**

  After adding `ScanAny*` / `PrintAny*` (#020), there was a duplication of logic between `scanVars*` and `scanAny*`, as well as between `printVals*`, `printWords*` and `printAny*`. 
  
  It is required that:
    - Develop a unified implementation that will allow the use of common loops for both typed and reflexive functions.
    - Keep the public API unchanged during implementation.
    - Evaluate the impact on performance using existing tests.
    - Decide whether to implement (or maintain the current approach)

## Made

- **007 [+] Adjust EOF handling in scanXxxLn functions (2026-03-11) (made:2026-03-11)**

  Currently, `scanSliceLn` and `scanVarsLn` may return `io.EOF` after successfully reading the last line if the input lacks a trailing newline. This forces users to explicitly check for `io.EOF` even after processing data. Change the behavior:
    - `scanSliceLn`: returns `nil` error if at least one token was read, even if EOF is encountered afterwards.
    - `scanVarsLn`: returns `nil` after reading all requested variables, ignore `io.EOF` when skipping the final spaces.

  This makes the API more convenient for typical contest usage, where EOF after data is not an error.

- **009 [+] Fix panic in Resize when expanding slice within capacity (2026-03-11) (made:2026-03-12)**

  **Problem:** The current implementation of `Resize` panics when `len(s) < n <= cap(s)` because it unconditionally calls `clear(s[n:])`.  
  **Solution:** Call `clear` only when `n < len(s)`.

  Move `xslices.go` under `sugar` build tag.

- **010 [+] Remove Sign and Unsig interfaces? (2026-03-12)(made:2026-03-16)**

  Currently, `Sign` and `Unsig` are only used to define `Int`. They are not used elsewhere in the library. Should we remove them and define `Int` directly as `~int | ~int8 | ... | ~uint | ...`? Removing would reduce public API surface, but might limit future extensibility. Decision needed.  

  I threw them away because I'm annoyed by these lines in the inlining.

- **011 [-] Replace custom type example with code generator? (2026-03-12) (made:2026-03-22)**

  The current example (`custom_type.go`) demonstrates how to create a custom type by copying and modifying library code. This requires forking the library, which is not ideal. Should we replace it with a `go generate` based generator that produces the necessary wrapper functions for any user-defined type? Or should we remove the example altogether and rely on the generic `ScanSlice`/`PrintSlice` (under sugar tag) as the primary way to handle custom types? The sugar approach already provides interfaces for custom types, but it's hidden behind a build tag and may have performance overhead. Decision needed.

  `custom_type.go` and `custom_type_sugar.go` have been removed.

- **012 [+] Verify compilation after inlining/clearing before replacing original file (2026-03-13) (made:2026-03-13)**

  **Problem:** After inlining library code (or clearing inlined code), the resulting `main.go` might become uncompilable (e.g., due to missing build tags or other issues). Previously, the tool would blindly overwrite the original file, leaving the user with a broken file. This is unacceptable — the file should remain in a working state or stay unchanged.

  **Solution:** Before renaming the backup to the original file, run a temporary `go build` with the provided `-tags` on the modified code. If the build fails, the operation aborts, preserving the original file. A new `--no-build-check` flag allows skipping this verification when needed. The check is performed both for `inline` and `clear` operations.

  Also improved formatting: remove empty lines between inlined declarations.

- **013 [+] Fix `generateInts` in `utils_test.go` to generate correctly length-distributed signed integers (2026-03-18) (made:2026-03-18)**

  **Problem:** The `generateInts` helper used in benchmarks produced numbers with incorrect length distribution for negative values — all negative values were long, skewing benchmark results.  
  **Solution:** Rewrite `generateInts` to generate random numbers with proper sign and uniform length distribution, ensuring realistic benchmarks.

- **014 [+] Unify parseInt implementations and remove build tags (made:2026-03-19)**

  **Problem:** The library previously offered three separate integer parsers (`std`, `base`, `fast`) selectable via build tags. This increased complexity and forced users to make an unnecessary choice.

  **Solution:** Replace with a single `parseInt` function that uses a fast manual loop for numbers shorter than 20 digits and falls back to `strconv.ParseUint` for longer inputs. This balances performance and simplicity, and is not worse than `strconv.Atoi` for typical contest use. All build tags and related code are removed; documentation is updated accordingly.

  This change simplifies the codebase without compromising practical performance.  Advanced techniques (e.g., SWAR) are being explored separately but are unlikely to be merged due to code size considerations.

- **015 [+] Override ReadBytes/ReadString in custom Reader/Writer (made:2026-03-19)**

  **Problem:** Standard `bufio.Reader.ReadBytes`/`ReadString` return trailing delimiters and don't trim whitespace, requiring extra code in contests. Also, they return `io.EOF` even when data was read, forcing boilerplate checks like `if err == io.EOF && len(data) > 0`.

  **Solution:** 
    - Introduce `contestio.Reader` and `Writer` that embed `*bufio.Reader` and `*bufio.Writer` respectively, preserving all original methods.
    - Override `ReadBytes` and `ReadString` to:
      * Remove the delimiter if present.
      * Trim trailing whitespace (space, tab, `\r`, `\n`).
      * Return `io.EOF` **only** when no data was read at all.
      * Otherwise ignore EOF and return the data (even if empty after trimming).
    - This behaviour aligns with typical contest input handling and eliminates manual EOF checks.

  **Impact:** All existing code using `bufio.Reader` directly must switch to `contestio.Reader` to benefit; otherwise, behaviour remains unchanged. The embedded types ensure backward compatibility for other methods like `Peek`, `Discard`, etc.

- **016 [+] Optional panic on scan/print errors via build tag `must` (made:2026-03-19)**

  **Problem:** In contest tasks, input is usually well-formed, so any I/O or parsing error is unexpected and should terminate the program. Manually writing `if err != nil { panic(err) }` after every scan/print call is repetitive and clutters the code.

  **Solution:** Introduce a new build tag `must`. When enabled (`-tags=must`), all public scan and print functions panic on any error except `io.EOF` (which is treated as normal end of input). The internal `must` helper wraps the common logic. Without the tag, functions return errors as before.

  This gives users a choice: explicit error handling or automatic panic for simpler code.

- **018 [+] Improve contestio-inline to accept directory argument (2026-03-21) (made:2026-03-22)**

  Currently, `contestio-inline` optionally accepts a file path; it defaults to `main.go` in the current directory. Enhance it to also accept a directory path: if the argument is a directory, look for `main.go` inside it. This will make it more convenient when working with projects where the main file is nested.

- **019 [+] Replace overridden ReadBytes/ReadString with ScanBytes/ScanString functions (2026-03-21) (made:2026-03-21)**

  Currently, `contestio.Reader` embeds `*bufio.Reader` and overrides `ReadBytes`/`ReadString` to provide trimmed tokens without delimiters and better EOF handling. This hides the original methods, which may be confusing and limiting for users who need the original behaviour.

  **Plan:**
    - Remove the overridden methods from `Reader`.
    - Keep `Reader` as an embedded `*bufio.Reader` (no method shadowing).
    - Provide package-level functions `ScanBytes(r *Reader, delim byte) ([]byte, error)` and `ScanString(r *Reader, delim byte) (string, error)` that implement the desired logic (trim spaces, drop delimiter, EOF only if no bytes read).

  This restores full access to `bufio.Reader` methods while keeping the convenience functions for token scanning. It also aligns with the library's pattern of using `ScanXXX` functions for higher-level operations.

- **020 [+] Mixed scan: two-stage implementation (2026-03-22) (made:2026-03-28)**

  Provide a convenient API for scanning mixed types (e.g., `ScanMixed(r, &name, &age, &score)`). 

  **Stage 1:** Implemented `ScanAny`, `ScanAnyLn`, `PrintAny`, `PrintAnyLn` under the `any` build tag.
  These functions accept `any` arguments and use reflection (or unsafe access to `eface` structure when built with `-tags=unsafe`) to provide a convenient API for mixed‑type I/O without code generation.
  The implementation is isolated behind the `any` tag to avoid polluting the main library with reflection overhead; users can enable it explicitly.

  **Stage 2:** Postponed. The reflection‑based API proved simpler and performant enough for typical contest usage; generating specialised functions for each call signature is not a priority at this time.
