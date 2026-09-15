// The design system's public surface. Screens and routes import from here;
// nothing reaches into src/ directly (.claude/rules/typescript.md).

// Atoms
export { default as Brand } from './atoms/Brand.svelte';
export { default as Icon } from './atoms/Icon.svelte';
export { default as Button } from './atoms/Button.svelte';
export { default as Notice } from './atoms/Notice.svelte';

// Molecules
export { default as EmptyState } from './molecules/EmptyState.svelte';
export { default as ErrorNotice } from './molecules/ErrorNotice.svelte';
export { default as InfoNotice } from './molecules/InfoNotice.svelte';
export { default as PendingButton } from './molecules/PendingButton.svelte';

// Organisms
export { default as SidebarNav } from './organisms/SidebarNav.svelte';

// Templates
export { default as AuthShell } from './templates/AuthShell.svelte';
export { default as SignedInLanding } from './templates/SignedInLanding.svelte';

// Pages — pure screens: data in as props, intent out as callbacks. They read
// nothing from $app, which is what lets them be storied without mocking.
export { default as EnrolScreen } from './pages/EnrolScreen.svelte';
export { default as LandingScreen } from './pages/LandingScreen.svelte';
export { default as LinkUnusableScreen } from './pages/LinkUnusableScreen.svelte';
export { default as SignInScreen } from './pages/SignInScreen.svelte';

export type { Destination } from './destinations.js';
export type { EnrolState } from './pages/EnrolScreen.svelte';
export type { SignInState } from './pages/SignInScreen.svelte';
