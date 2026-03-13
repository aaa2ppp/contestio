# Notes on the "Contest IO" Project

<!-- next-note-id:013 -->

## Open Questions

- **004 [ ] Public SkipSpace/SkipSpaceLn/SkipLn (2026-03-09)**

  Should we make these functions public?
  Currently, this functionality is implemented via private functions `skipSpace`/`skipToNewLine` (used inside `scanXXX`).  
  Users may want to explicitly use them in their code, but I haven't encountered any such cases yet.

- **005 [ ] **Move library code out of root?** (2026-03-09)**

  Currently all library files are in the root. Should we move them to a subdirectory to keep root clean?  
  If so, what name? I dislike plain `lib`. Perhaps just keep as is and move examples to `examples/`?

- **006 [ ] What to do with Sugar? (2026-03-10)**

  The `ScanSlice` and `PrintSlice` functions accept `Parser` and `ValAppender` interfaces, respectively.
  This allows scanning a slice of any type using only these functions.
  The solution seems heavyweight (performance overhead) and clunky (requires implementing an interface).
  Currently, this functionality is isolated from the main codebase with the build tag `-tags=sugar`.
  Should we delete it or merge it into the main codebase?

- **008 [ ] Should nextToken/scanXxx advance position on read/parse errors? (2026-03-11)**

  Currently, `nextToken` always discards the bytes it has consumed, even if it returns an error (e.g., `ErrTokenTooLong` or `io.EOF` after a partial token). Similarly, `scanSlice`/`scanVars` advance the reader past the token that caused a parse error before returning the error. This behavior is undocumented.

  **Pros of current approach:**
  - Simple and consistent: after an error, the stream is left at the next possible position.
  - Useful for resuming after a recoverable error (e.g., skipping malformed input).

  **Cons:**
  - If the caller wants to inspect or retry the erroneous token, the data is lost.
  - Error recovery becomes non-trivial because the state is already mutated.

  Should we change the behavior to **not advance** on parse errors (i.e., leave the reader pointing at the beginning of the offending token)? Or should we keep the current behavior but document it clearly? This decision affects all scanning functions and the `nextToken` helper.

- **010 [ ] Remove Sign and Unsig interfaces? (2026-03-12)**

  Currently, `Sign` and `Unsig` are only used to define `Int`. They are not used elsewhere in the library. Should we remove them and define `Int` directly as `~int | ~int8 | ... | ~uint | ...`? Removing would reduce public API surface, but might limit future extensibility. Decision needed.  


- **011 [ ] Replace custom type example with code generator? (2026-03-12)**

  The current example (`custom_type.go`) demonstrates how to create a custom type by copying and modifying library code. This requires forking the library, which is not ideal. Should we replace it with a `go generate` based generator that produces the necessary wrapper functions for any user-defined type? Or should we remove the example altogether and rely on the generic `ScanSlice`/`PrintSlice` (under sugar tag) as the primary way to handle custom types? The sugar approach already provides interfaces for custom types, but it's hidden behind a build tag and may have performance overhead. Decision needed.

  
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

- **012 [+] Verify compilation after inlining/clearing before replacing original file (2026-03-13) (made:2026-03-13)**

  **Problem:** After inlining library code (or clearing inlined code), the resulting `main.go` might become uncompilable (e.g., due to missing build tags or other issues). Previously, the tool would blindly overwrite the original file, leaving the user with a broken file. This is unacceptable — the file should remain in a working state or stay unchanged.

  **Solution:** Before renaming the backup to the original file, run a temporary `go build` with the provided `-tags` on the modified code. If the build fails, the operation aborts, preserving the original file. A new `--no-build-check` flag allows skipping this verification when needed. The check is performed both for `inline` and `clear` operations.

  Also improved formatting: remove empty lines between inlined declarations.
