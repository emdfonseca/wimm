<script lang="ts">
	import type { Snippet } from 'svelte';
	import Notice from '../atoms/Notice.svelte';

	/**
	 * Notice · info. Polite by default: "you closed the prompt, nothing
	 * happened" and "you were signed out" both find the member returning from a
	 * native prompt, and interrupting them loses their place.
	 */
	interface Props {
		title?: string;
		live?: 'assertive' | 'polite' | 'off';
		children?: Snippet;
	}

	let { title, live = 'polite', children }: Props = $props();
</script>

<div class="tone">
	<Notice {title} {live}
		>{#if children}{@render children()}{/if}</Notice
	>
</div>

<style>
	.tone {
		--notice-bg: var(--color-feedback-info-bg);
		--notice-accent: var(--color-feedback-info);
	}
</style>
