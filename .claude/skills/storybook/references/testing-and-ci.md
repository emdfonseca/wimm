# Testing and CI

## Stories are the test corpus

With `@storybook/addon-vitest`, every story runs as a test: it must render without error, its play function must pass, and its a11y checks must pass at whatever level `parameters.a11y.test` sets.

Installing the Vitest addon into a project that already has `@storybook/addon-a11y` wires the accessibility annotations into `.storybook/vitest.setup.ts`. Verify that file exists after setup — if it does not, a11y checks silently do not run in the test pass, which looks identical to passing.

## just recipes

Recipes live in the package justfile (`monorepo`): `just dev packages/ui` starts Storybook, `just check packages/ui` runs the story tests. No separate `just storybook` verb.

## CI

`devbox run -- just check packages/ui` covers it. Add once the design system stabilises:

- **Build the static Storybook** (`storybook build`) — catches broken stories the test runner tolerates and gives a deployable URL for design review.
- **Visual regression** (Chromatic or equivalent) only when there are enough components that visual drift is otherwise unnoticeable.

## What Storybook does not test

Routing, data loading, authentication, and anything spanning more than one screen. Those are flows across wired screens, and they need e2e tests against the running app.
