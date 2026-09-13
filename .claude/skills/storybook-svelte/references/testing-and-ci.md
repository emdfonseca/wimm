# Testing and CI

## Stories are the test corpus

With `@storybook/addon-vitest`, every story runs as a test: it must render without error, its play function must pass, and its a11y checks must pass at whatever level `parameters.a11y.test` sets. That is a large amount of coverage for artifacts that already had to exist.

Installing the Vitest addon into a project that already has `@storybook/addon-a11y` wires the accessibility annotations into `.storybook/vitest.setup.ts` automatically. Verify that file exists after setup — if it does not, a11y checks silently do not run in the test pass, which looks identical to passing.

## just verbs

Storybook is part of the design-system package, so it uses the standard verb set:

```just
# packages/ui/justfile
dev:
    pnpm storybook dev -p 6006

build:
    pnpm build
    pnpm storybook build

test:
    pnpm vitest run

lint:
    pnpm eslint .

check: lint
    pnpm svelte-check
    just test
```

No separate `just storybook` verb. A newcomer running `just dev packages/ui` should get the workshop, and `just check packages/ui` should run everything that can fail.

## CI

`devbox run -- just check packages/ui` covers it. Two things worth adding once the design system stabilises:

- **Build the static Storybook** in CI (`storybook build`) — it catches broken stories that the test runner tolerates, and gives reviewers a deployable URL for design review.
- **Visual regression** (Chromatic or an equivalent) if pixel drift matters to the team. Worth saying plainly: this has real cost in review time and flake, and it pays off only when there are enough components that visual drift is otherwise unnoticeable. Adopt it when that is true, not before.

## What Storybook does not test

Routing, data loading, authentication, and anything spanning more than one screen. Those are journeys, they live in the `.pen` files, and they need e2e tests against the running app. Keeping that line sharp is what stops Storybook from slowly turning into a slow, mock-heavy integration suite that verifies mocks.
