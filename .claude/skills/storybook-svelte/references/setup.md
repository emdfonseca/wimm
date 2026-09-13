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

Pin the toolchain (node, pnpm) in `devbox.json` as usual; Storybook itself is a normal dev dependency of the design-system package.

## main.ts

```ts
import type { StorybookConfig } from '@storybook/sveltekit';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.svelte', '../src/**/*.mdx'],
  framework: '@storybook/sveltekit',
  addons: [
    '@storybook/addon-svelte-csf',
    '@storybook/addon-docs',
    '@storybook/addon-a11y',
    '@storybook/addon-vitest',
  ],
};

export default config;
```

Storybook belongs to the design-system package (`packages/ui/.storybook/`), not the repo root — it documents that package, and keeping it there means `just dev packages/ui` starts it.

## preview.ts

Three jobs: import the real token stylesheet, declare the theme global, declare the viewport presets.

```ts
import type { Preview } from '@storybook/sveltekit';
import '../src/lib/styles/tokens.css';   // the real tokens, not a copy

const preview: Preview = {
  parameters: {
    a11y: { test: 'error' },
    viewport: {
      options: {
        compact: { name: 'Compact (390)', styles: { width: '390px', height: '844px' }, type: 'mobile' },
        medium:  { name: 'Medium (768)',  styles: { width: '768px', height: '1024px' }, type: 'tablet' },
        wide:    { name: 'Wide (1440)',   styles: { width: '1440px', height: '900px' }, type: 'desktop' },
      },
    },
  },
  initialGlobals: {
    theme: 'light',
    viewport: { value: 'wide', isRotated: false },
  },
  globalTypes: {
    theme: {
      description: 'Color theme',
      toolbar: {
        title: 'Theme',
        icon: 'circlehollow',
        items: [
          { value: 'light', icon: 'sun', title: 'Light' },
          { value: 'dark', icon: 'moon', title: 'Dark' },
        ],
        dynamicTitle: true,
      },
    },
  },
  decorators: [
    (story, { globals }) => {
      document.documentElement.dataset.theme = globals.theme ?? 'light';
      return story();
    },
  ],
};

export default preview;
```

`@storybook/addon-themes` (`withThemeByDataAttribute`) does the same job if you prefer an addon to a hand-written decorator; either is fine, but pick one — two theme mechanisms fighting over the same attribute is a confusing afternoon.

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

Needing heavy mocking is a design signal: a library component reaching into `$app/state` is coupled to the app. Push that dependency up into the route and pass data as props — the component gets simpler and the story stops needing a fixture.
