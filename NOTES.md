# Notes on the "Contest IO" Project

<!-- next-note-id:006 -->

## Open Questions

- **004 [ ] Public SkipSpace/SkipSpaceLn/SkipLn (2026-03-09)**

  Should we make these functions public?
  Currently, this functionality is implemented via private functions `skipSpace`/`skipToNewLine` (used inside `scanXXX`).  
  Users may want to explicitly use them in their code, but I haven't encountered any such cases yet.

- **005 [ ] **Move library code out of root?** (2026-03-09)**

  Currently all library files are in the root. Should we move them to a subdirectory to keep root clean?  
  If so, what name? I dislike plain `lib`. Perhaps just keep as is and move examples to `examples/`?

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
