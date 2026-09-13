# TypeScript

## Vitest, beside the code

`foo.test.ts` next to `foo.ts`. `describe` per unit, `it` per behaviour named as a sentence:

```ts
describe('formatMoney', () => {
  it('renders cents as a two-decimal amount', () => {
    expect(formatMoney({ cents: 1050, currency: 'EUR' })).toBe('€10.50');
  });
  it('rejects a negative amount', () => {
    expect(() => formatMoney({ cents: -1, currency: 'EUR' })).toThrow(RangeError);
  });
});
```

Run one while cycling: `pnpm vitest run -t 'rejects a negative amount'`.

## What is a unit test here, and what is not

Pure functions, stores, data transformations, client-side validation, anything in `src/lib` that does not render: unit test with TDD.

Svelte components: the visual states are Storybook stories and the behavioural contracts are play functions (`storybook-svelte`). A component's *logic* that is worth TDD is usually logic that wants to be extracted into a plain function anyway — do that, test the function, and let the story cover the rendering.

Route modules (`+page.server.ts` load functions, form actions): thin by construction; the domain call they make is what gets tested. If a load function has enough logic to want a test, move the logic out.

## Doubles

`vi.fn()` to observe a callback fired. In-memory fakes for stores and clients — a `Map`-backed implementation of the small interface the code needs. `vi.mock` of whole modules is the last resort; it couples the test to import structure.

## Time and randomness

`vi.useFakeTimers()` with `vi.setSystemTime()`; inject ID generators. No real timers, no `Date.now()` in code under test.

## Generated clients

`@repo/contracts` types are the contract; tests use them, never hand-written shapes. A test that constructs a request the generated type would reject is testing the wrong thing.
