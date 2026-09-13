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
