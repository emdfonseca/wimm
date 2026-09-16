<script lang="ts">
	import type { Snippet } from 'svelte';
	import Notice from '../atoms/Notice.svelte';

	/**
	 * Notice · error. Assertive by default: "your device did not save a passkey"
	 * and "that passkey is not enrolled here" both stop the member getting what
	 * they came for, so they interrupt rather than wait.
	 */
	interface Props {
		title?: string;
		live?: 'assertive' | 'polite' | 'off';
		children?: Snippet;
	}

	let { title, live = 'assertive', children }: Props = $props();
</script>

<div class="tone">
	<Notice {title} {live}
		>{#if children}{@render children()}{/if}</Notice
	>
</div>

<style>
	.tone {
		--notice-bg: var(--color-feedback-error-bg);
		--notice-accent: var(--color-feedback-error);
	}
</style>
