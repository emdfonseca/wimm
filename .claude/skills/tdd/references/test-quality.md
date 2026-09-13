# What makes a test worth keeping

A test earns its place by failing when the behaviour it names breaks, and only then. Most bad tests fail one of those two halves: they never fail (tautological, asserting on mocks), or they fail for unrelated reasons (brittle, coupled to structure).

## One behaviour, named as a sentence

`TestInvoice_RejectsNegativeTotal`, `it('returns the original response on idempotent replay')`. The name states the rule; a failure reads as "this rule is broken". A test named `TestNew` or `it('works')` forces the reader into the body to learn what broke.

One assertion cluster per test. Two behaviours in one test means a failure in the first hides the second.

## Arrange, act, assert — visibly

Three short blocks, in that order, no logic. A test with an `if` or a loop is a program that itself needs testing. Table-driven tests are the exception that proves the rule: the loop is the harness, and each row is still one behaviour.

## Fakes over mocks, interfaces at the consumer

A fake is a small working implementation (in-memory store, recording clock). A mock is a script of expected calls. Fakes test *outcomes*; mocks test *choreography*, and choreography changes every refactor. Prefer fakes, and define the interface the code under test needs — three methods, in the consumer's package — rather than mocking a 30-method client.

Use `fn()` / `vi.fn()` / `MagicMock` to *observe* a call was made when the call is the outcome (a notification was sent), not to drive the whole test.

## Deterministic and fast

Inject the clock, the random source, and IDs. No `time.Sleep` to wait for things; no network; no shared mutable state between tests. A test that passes on retry is a test that is failing intermittently, and it will be ignored within a month.

## Don't test private structure

Test through the package's public surface. A test that reaches into unexported functions locks the current decomposition in place — exactly the thing the refactor step is supposed to be free to change. If something private is hard to reach through the public API, that is often a sign it wants to be its own small package.

## The mutation check

Before committing, break the implementation deliberately — flip the condition, return the wrong value — and run the test. It must fail. If it does not, the test is not testing what you think. This takes ten seconds and catches the majority of tautological tests.

## Assertions on generated output

Golden/approval files for large structured output (rendered templates, serialised contracts). Commit the golden file; update it deliberately with a flag (`-update`), and read the diff when you do. A golden file updated without reading the diff is a test that asserts "whatever it does now".
