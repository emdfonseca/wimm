// The one declaration of the viewport sizes. Storybook's toolbar reads it
// through preview.ts and the design canvas reads it directly, so the two
// cannot disagree. Plain ESM because a static page cannot import TypeScript.
export const viewportOptions = {
	compact: { name: 'Compact', styles: { width: '390px', height: '844px' } },
	medium: { name: 'Medium', styles: { width: '834px', height: '1112px' } },
	wide: { name: 'Wide', styles: { width: '1440px', height: '900px' } },
	ultra: { name: 'Ultra', styles: { width: '1920px', height: '1080px' } }
};
