import type { Preview } from '@storybook/sveltekit';
import '../src/lib/styles/tokens.css';

const preview: Preview = {
  parameters: {
    a11y: { test: 'error' },
    controls: { matchers: { color: /(background|color)$/i, date: /Date$/ } },
    viewport: {
      options: {
        compact: { name: 'Compact (390)', styles: { width: '390px', height: '844px' }, type: 'mobile' },
        medium: { name: 'Medium (768)', styles: { width: '768px', height: '1024px' }, type: 'tablet' },
        wide: { name: 'Wide (1440)', styles: { width: '1440px', height: '900px' }, type: 'desktop' },
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
