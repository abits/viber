---
name: go-best-practices
description: Idiomatic Go rules distilled from Effective Go plus modern (post-2018) additions — naming, errors, concurrency, context, testing, generics. Use when writing, reviewing, or refactoring Go code in this repo.
---

Every rule below is an imperative — apply it mechanically unless a rule higher in this list overrides it.

**Version floor:** current stable Go (assume ≥ 1.22 unless the repo's `go.mod` says otherwise). Features requiring a specific minimum are annotated inline.

**Precedence:** if `gofmt`, `go vet`, or the repo's `golangci-lint` config disagrees with anything here, the tool wins. This file is guidance, not law.

## 1. Formatting

- Use `gofmt` / `goimports`. Never format by hand and never argue about tabs, brace placement, or import grouping — the linter and the format hook already agreed.

## 2. Naming

- Package names: lowercase, one word, no underscores, no `MixedCaps`. Match the directory.
- Exported = `MixedCaps`; unexported = `mixedCaps`. Never `snake_case` in Go.
- No stuttering with the package name: `bytes.Buffer`, not `bytes.BytesBuffer`; `http.Request`, not `http.HTTPRequest`.
- Getters drop the `Get`: `owner.Name()`, not `owner.GetName()`. Setters keep `Set`: `owner.SetName(...)`.
- Single-method interfaces take the `-er` suffix: `Reader`, `Writer`, `Closer`, `Stringer`.
- Canonical method names have canonical signatures. Do not invent variants.

  ```go
  // GOOD
  String() string
  Error() string
  Read(p []byte) (n int, err error)
  Write(p []byte) (n int, err error)
  Close() error
  Len() int

  // BAD — invented variant of a canonical name
  ToString() string
  ErrorMessage() string
  WriteData(b []byte) error
  ```

- Receiver names are 1–2 letters and consistent across every method of the same type. Not `this`, not `self`.

  ```go
  // GOOD — every method on *Server uses s
  func (s *Server) Start() error   { /* ... */ }
  func (s *Server) Stop(ctx context.Context) error { /* ... */ }

  // BAD — mixed receiver names
  func (srv *Server) Start() error { /* ... */ }
  func (s *Server)   Stop() error  { /* ... */ }
  ```

## 3. Control flow

- `for` is the only loop. No `while`, no `do-while`.

  ```go
  for cond { }          // while
  for { }               // infinite
  for i := 0; i < n; i++ { }
  for k, v := range m { }
  ```

- Braces are mandatory even for single-statement bodies.
- Drop `else` after a terminal statement (`return`, `break`, `continue`, `panic`).

  ```go
  // GOOD
  if err != nil {
      return err
  }
  doNext()

  // BAD
  if err != nil {
      return err
  } else {
      doNext()
  }
  ```

- Use the initializer form of `if` and `switch` to keep temporaries scoped:

  ```go
  if v, ok := m[k]; ok {
      use(v)
  }
  ```

- `switch` cases do not fall through. Add `fallthrough` on its own line only when you really mean it.
- Type switch is `switch v := x.(type)`. It only works inside a `switch`.
- `range` semantics:
  - Over a string → `(index, rune)`. Use `[]byte(s)` if you need bytes.
  - Over a map → random order. Sort keys explicitly if you need determinism.
  - Over a channel → until closed. Sender must close.

## 4. Data

- `new(T)` returns `*T` with the zero value. Useful only when the zero value is meaningful (`new(bytes.Buffer)`, `new(sync.Mutex)`).
- `make(T, args)` is for slices, maps, and channels only. It returns `T`, not `*T`.
- `new([]int)` is almost never right — use `make([]int, 0, n)` or `var s []int`.
- Composite literals use field names. Never rely on positional order for structs with more than two fields — it breaks silently when the type gains a field.

  ```go
  // GOOD
  p := Point{X: 1, Y: 2}

  // BAD — order-dependent
  p := Point{1, 2}
  ```

- `append` may reallocate. Always assign the result.

  ```go
  s = append(s, x)   // GOOD
  append(s, x)       // BAD — silent no-op if it reallocates
  ```

- Slice aliasing: `s[a:b]` shares the backing array with `s`. If the caller mustn't see mutations, `copy` into a fresh slice.
- Nil map: reads return the zero value; writes panic. Nil slice: reads and `append` are both fine.
- Two-dimensional slices need per-row allocation — a single `make([][]int, n)` gives `n` nil rows.

## 5. Functions & methods

- Multi-return with `error` last. Never invent out-parameters or error codes.
- Named return values are for godoc clarity, not for cleverness. Do not lean on them to skip explicit `return` statements in long functions — it hides the flow.
- Pick value vs pointer receivers per type and stay consistent across every method of that type. Pointer receiver is required for:
  - mutation of the receiver,
  - large structs (avoid the copy),
  - any interface implementation that exposes mutating methods.
- `defer` runs LIFO at function return. Beware in loops — deferred closes stack up until the enclosing function ends.

  ```go
  // BAD — every iteration accumulates a deferred Close until the outer function returns
  for _, path := range paths {
      f, err := os.Open(path)
      if err != nil {
          return err
      }
      defer f.Close()
      process(f)
  }

  // GOOD — scope the defer to one iteration
  for _, path := range paths {
      err := func() error {
          f, err := os.Open(path)
          if err != nil {
              return err
          }
          defer f.Close()
          return process(f)
      }()
      if err != nil {
          return err
      }
  }
  ```

## 6. Interfaces

- Go interfaces are implicit. No `implements` keyword. A type satisfies an interface as soon as it has the methods.
- Keep interfaces small — usually 1–3 methods. Composition beats a wide interface.
- Compile-time assertion pattern for stability:

  ```go
  var _ io.Reader = (*MyReader)(nil)
  ```

- Interfaces belong to the *consumer*, not the producer. Declare them in the package that uses them.
- Accept an interface, return a concrete type. This lets callers substitute at the boundary while your package documents exactly what it produces.
- Embedding is composition, not inheritance. Methods of an embedded type are *promoted*; there is no `super`, no virtual dispatch, no override.

## 7. Errors (this is where LLMs go wrong most often)

- Return `error`; do not `panic` for expected failure modes.
- Wrap with `%w` to add context and preserve the chain:

  ```go
  if err := os.WriteFile(path, buf, 0o644); err != nil {
      return fmt.Errorf("write %s: %w", path, err)
  }
  ```

- Inspect with `errors.Is` (sentinel) and `errors.As` (typed). Never `==` after wrapping is possible.

  ```go
  // GOOD
  if errors.Is(err, os.ErrNotExist) { /* ... */ }
  var pathErr *fs.PathError
  if errors.As(err, &pathErr) { /* use pathErr */ }

  // BAD — misses any wrapped error
  if err == os.ErrNotExist { /* ... */ }
  ```

- Sentinel errors are package-level `var`s named `ErrX`:

  ```go
  var ErrNotFound = errors.New("not found")
  ```

- Typed errors are structs with an `Error() string` method and inspectable fields:

  ```go
  type ValidationError struct {
      Field string
      Msg   string
  }
  func (e *ValidationError) Error() string {
      return fmt.Sprintf("validation: %s: %s", e.Field, e.Msg)
  }
  ```

- `errors.Join(a, b, ...)` (Go 1.20+) combines multiple errors when a batch operation partially fails. `errors.Is`/`As` traverse the join tree.
- Never silently discard an error. `_ = f()` is only acceptable with a comment explaining why the failure is safe to ignore.
- `panic`/`recover` are for genuinely unrecoverable conditions (unreachable code, package init failure) or for stopping a goroutine that would otherwise crash the process. Not for control flow.

## 8. Concurrency

- Design principle: "share memory by communicating." Channels + goroutine ownership scale as designs; mutex + shared state scales as bugs.
- Every goroutine needs a termination story: `context.Context` cancellation, a closed channel, or a `sync.WaitGroup`. Leaked goroutines are latent bugs — the runtime will not report them.
- **Sender closes** the channel; receiver reads with `v, ok := <-ch` to detect close. Never close from the receiver side.
- `for v := range ch` reads until the channel is closed.
- `select` for multiple channels — always include an exit case:

  ```go
  select {
  case v := <-ch:
      handle(v)
  case <-ctx.Done():
      return ctx.Err()
  }
  ```

- Buffered channels (`make(chan T, n)`) are a bounded queue, not a synchronization primitive. Do not use buffering to "fix" a leak.
- `sync.Mutex` is fine when channelizing the state genuinely complicates the design. Keep the mutex next to the field it protects; embed only if all fields need the same lock.

  ```go
  type counter struct {
      mu sync.Mutex
      n  int
  }
  ```

- `sync.WaitGroup`: `wg.Add(n)` **before** the `go` statement, never inside the goroutine (race).

  ```go
  var wg sync.WaitGroup
  for _, item := range items {
      wg.Add(1)
      go func(it Item) {
          defer wg.Done()
          process(it)
      }(item)
  }
  wg.Wait()
  ```

- `sync.Once` for lazy one-time initialization of a singleton.
- `sync/atomic.Pointer[T]`, `atomic.Int64`, etc. (Go 1.19+) for type-safe atomics. Skip `unsafe.Pointer` gymnastics.
- Loop-variable capture: Go 1.22+ scopes each iteration. On older Go, the classic bug bites — shadow inside the loop:

  ```go
  // BAD on Go <1.22
  for _, v := range xs {
      go func() { use(v) }()
  }

  // GOOD everywhere
  for _, v := range xs {
      v := v
      go func() { use(v) }()
  }
  ```

## 9. Context

- `context.Context` is the first parameter, always named `ctx`.
- Never store a `Context` in a struct. Pass it through the call chain.
- Create derived contexts with `context.WithTimeout`, `WithDeadline`, or `WithCancel` — always `defer cancel()`.

  ```go
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()
  ```

- Long-running or I/O-blocking code must observe cancellation:

  ```go
  for {
      select {
      case <-ctx.Done():
          return ctx.Err()
      case msg := <-in:
          handle(msg)
      }
  }
  ```

- `context.Background()` at top-level entry points (`main`, HTTP handlers, test setup).
- `context.TODO()` when the surrounding code hasn't plumbed a context yet — placeholder, not a permanent choice.
- Do not wrap `context.Context` in a custom type unless there is a hard requirement. It defeats the standard middleware ecosystem.

## 10. Modules & packages

- `go.mod` at the module root. Import paths are module-relative.
- Code under `internal/` is unimportable outside the module. Anything above `internal/` is public API — treat renames, signature changes, and removals accordingly.
- Every package gets exactly one file with a package doc comment:

  ```go
  // Package steps runs the ordered init pipeline and reports progress.
  package steps
  ```

- Circular imports are a compile error and a design smell. Break the cycle by extracting the shared types into a lower-level package, or by using an interface defined in the consumer.
- Run `go mod tidy` before every commit. Vendored deps? Only with a very good reason.

## 11. Testing

- Table-driven tests are the default shape:

  ```go
  func TestAdd(t *testing.T) {
      cases := []struct {
          name    string
          a, b, want int
      }{
          {"zero", 0, 0, 0},
          {"positive", 2, 3, 5},
          {"mixed", -2, 5, 3},
      }
      for _, tc := range cases {
          t.Run(tc.name, func(t *testing.T) {
              if got := Add(tc.a, tc.b); got != tc.want {
                  t.Errorf("Add(%d,%d)=%d want %d", tc.a, tc.b, got, tc.want)
              }
          })
      }
  }
  ```

- `t.Helper()` in any shared assertion helper so failures point at the caller, not the helper.
- `t.Cleanup(func() { ... })` for teardown. Runs LIFO after the test even on failure.
- `t.TempDir()` gives a scoped, auto-cleaned directory. Never `os.TempDir()` + hand-roll cleanup.
- `t.Parallel()` inside sub-tests when they're independent. On Go <1.22, shadow the loop var.
- Prefer in-memory fakes: `testing/fstest.MapFS` for filesystem, `net/http/httptest.Server` for HTTP.
- Stick with the standard `testing` package. Add `testify/require` only when it clarifies materially — do not import it just to write `require.Equal(t, got, want)` in place of an `if`.

## 12. Generics (Go 1.18+)

- Reach for type parameters when the same logic applies to multiple concrete types and you'd otherwise duplicate code or use `any` + reflection.
- Constraints: `any` (any type), `comparable` (== / != valid), or a custom interface listing methods or approximating underlying types with `~T`.

  ```go
  type Number interface {
      ~int | ~int64 | ~float32 | ~float64
  }

  func Sum[T Number](xs []T) T {
      var s T
      for _, x := range xs {
          s += x
      }
      return s
  }
  ```

- Don't over-generalize. A duplicated 10-line function is clearer than a generic 30-line one with constraint gymnastics.

## 13. Modern standard library

- `any` in place of `interface{}` in new code (Go 1.18+).
- Builtins: `min`, `max`, `clear` (Go 1.21+). Don't reimplement.
- `slices` package (Go 1.21+): `Contains`, `Sort`, `SortFunc`, `Equal`, `Concat`, `Delete`, `Index`, `Reverse`.
- `maps` package (Go 1.21+): `Copy`, `Equal`, `Keys` (returns `iter.Seq` in 1.23+).
- `strings.Builder` for repeated concatenation; `bytes.Buffer` for bytes. Never `s += ...` in a loop.
- `sync/atomic.Int64`, `atomic.Uint64`, `atomic.Bool`, `atomic.Pointer[T]` — typed atomics.
- `time.Now().UTC()` for logs and anything serialized. Never store local time on the wire.

## 14. Anti-patterns to reject on sight

Reject or refactor when reviewing generated code:

- `GetFoo()` / `SetFoo()` for simple accessors → drop the `Get`.
- `err == someErr` after any chance of wrapping → `errors.Is(err, someErr)`.
- `interface{}` in new code → `any`.
- Writing to a nil map without `make` first.
- `defer f.Close()` inside a tight loop without a wrapping `func()`.
- `go func() { use(v) }()` inside a range on Go <1.22 without shadowing `v`.
- `panic` inside library code (as opposed to `main` or top-level init).
- Bare `return err` at every layer with no added context — wrap with a short operation name (`fmt.Errorf("<op>: %w", err)`) wherever it helps locate the failure.
- `context.Context` stored on a struct.
- Hand-rolled `Contains` / `Copy` / `min` / `max` when the stdlib helper exists.
- `_ = err` without a comment explaining why the error is safe to drop.

## 15. Build tags & cross-compilation

- `//go:build linux && amd64` on line 1 of the file, followed by a blank line, then `package`. Never mix with the deprecated `// +build` syntax.
- File-name suffixes are automatic constraints: `foo_linux.go`, `foo_amd64.go`, `foo_test.go`, `foo_linux_amd64.go`.
- Keep platform-specific code in its own file. Do not try to `runtime.GOOS ==` around large blocks — split them.

## 16. When in doubt

- Run `go vet ./...` and `golangci-lint run ./...`. If they flag it, fix it before shipping.
- Consult `go doc <pkg>` (or `pkg.go.dev`) before inventing something new.
- Cargo-cult from the standard library: `net/http`, `io`, `encoding/json`, `database/sql` are the canonical style. If you're unsure how something should look in Go, read the equivalent shape in one of those packages first.
