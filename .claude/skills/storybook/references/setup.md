# Setup

API current as of Storybook 10.2 docs, checked 2026-09-13. Verify against current docs if behaviour differs.

## Packages

```text
@storybook/sveltekit          framework (use this, not @storybook/svelte, in a SvelteKit app)
@storybook/addon-svelte-csf   write stories as .stories.svelte
@storybook/addon-a11y         accessibility checks
@storybook/addon-vitest       run stories as tests
@storybook/addon-docs         autodocs / MDX pages
```

Storybook is added as a dev dependency of `packages/ui`; toolchain versions come from `devbox.json`.

## main.ts

`npx storybook init` in `packages/ui` generates `.storybook/main.ts`. Confirm it has `framework: '@storybook/sveltekit'`, the five addons above, and `stories: ['../src/**/*.stories.svelte', '../src/**/*.mdx']`. Storybook lives in `packages/ui/.storybook/`, not the repo root.

## preview.ts

Copy `assets/preview.ts` to `packages/ui/.storybook/preview.ts`. It imports the real token stylesheet (`../src/lib/styles/tokens.css`), sets `a11y.test: 'error'`, declares the `theme` global with a `data-theme` decorator, and declares the `compact` / `medium` / `wide` viewport presets.

`@storybook/addon-themes` (`withThemeByDataAttribute`) does the same job as the decorator; pick one mechanism, never both.

## SvelteKit mocking

Components that touch SvelteKit runtime modules need those mocked per story, via `parameters.sveltekit_experimental`:

```ts
parameters: {
  sveltekit_experimental: {
    // $app/navigation
    navigation: {
      goto: (url) => console.log('goto', url),
      invalidateAll: () => {},
    },
    // $app/forms
    forms: { enhance: () => {} },
    // anchor clicks
    hrefs: {
      '/settings': (to) => console.log('navigate', to),
      '/docs.*': { callback: (to) => {}, asRegex: true },
    },
    // $app/state (or $app/stores)
    state: {
      page: { data: { user: { name: 'Ada' } } },
      navigating: null,
    },
  },
}
```

A library component reaching into `$app/state` is coupled to the app: push that dependency into the route and pass data as props.
