<script lang="ts">
	import '@wimm/ui/base.css';
	import { navigating, page } from '$app/state';
	import { NavigationProgress } from '@wimm/ui';
	import { navigationAnnouncement } from './navigation-announcement.js';

	let { children } = $props();

	const announcement = $derived(navigationAnnouncement(Boolean(navigating.to), page.route.id));
</script>

<NavigationProgress active={Boolean(navigating.to)} />
<p class="sr-only" role="status" aria-live="polite">{announcement}</p>

{@render children()}

<style>
	.sr-only {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		margin: -1px;
		padding: 0;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
</style>
