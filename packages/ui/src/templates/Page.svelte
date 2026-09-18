<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * Origin `BPW3V`. Sidebar, page header, optional secondary nav, one content
	 * slot — build exactly this. It owns the page gutter, the section gap, and
	 * the heading's type; the six hand-rolled `.screen` rules and the four
	 * disagreeing `h1` blocks it replaces each drew their own copy of this.
	 *
	 * `sections` is a subsection list, per ADR 0005: a card in the page at
	 * every regime, never a full-height column beside the sidebar. At Compact
	 * `layout-subnav-width` is 0 and it stacks above the content rather than
	 * beside it.
	 *
	 * `inspector` carries the 400px region the library's `Drawer` describes at
	 * Ultra. No screen in wimm opens a drawer yet, so nothing passes it —
	 * recorded as a deliberate gap in this change's canvas rather than left
	 * unexplained.
	 *
	 * `layout-content-max` is not applied here: it caps text measure only, and
	 * a panel, a table or a ledger takes the full content column (ADR 0005).
	 * A screen wraps its own prose in it where that applies.
	 */
	interface Props {
		title: string;
		/** The page's own single primary action, in the header beside the title. */
		action?: Snippet;
		sections?: Snippet;
		inspector?: Snippet;
		/** Moves focus to the heading on mount. A member reaching this screen by
		 *  a full navigation — coming back from a bank, say — lands where a
		 *  screen reader has already announced the page, and without this has
		 *  no way to know the navigation landed. Every screen that needs it drew
		 *  its own copy of the same effect; this is the one place it lives now. */
		focusHeading?: boolean;
		children: Snippet;
	}

	let { title, action, sections, inspector, focusHeading = false, children }: Props = $props();

	let headingElement = $state<HTMLHeadingElement | null>(null);

	$effect(() => {
		if (focusHeading) headingElement?.focus();
	});
</script>

<div class="page">
	<header class="header">
		<h1 bind:this={headingElement} tabindex="-1">{title}</h1>
		{#if action}
			<div class="action">{@render action()}</div>
		{/if}
	</header>

	<div class="body">
		{#if sections}
			<nav class="sections" aria-label="Sections">{@render sections()}</nav>
		{/if}

		<div class="content">
			{@render children()}
		</div>

		{#if inspector}
			<aside class="inspector">{@render inspector()}</aside>
		{/if}
	</div>
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-block-size: 100%;
		inline-size: 100%;
	}

	.header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		flex: none;
		block-size: var(--layout-header-height);
		padding-inline: var(--layout-page-gutter);
		background: var(--color-bg-surface);
		border-block-end: 1px solid var(--color-border-default);
	}

	h1 {
		margin: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-primary);
		font-family: var(--type-family-display);
		font-size: var(--type-size-page-title);
		font-weight: 600;
	}
	/* No ring: the heading takes focus to place a screen reader, not to show a
	   keyboard user where they are. */
	h1:focus,
	h1:focus-visible {
		outline: none;
	}

	.body {
		display: flex;
		flex: 1;
		min-block-size: 0;
		gap: 24px;
		padding: var(--layout-page-gutter);
	}

	.sections {
		flex: none;
		inline-size: var(--layout-subnav-width);
	}

	.content {
		display: flex;
		flex-direction: column;
		flex: 1;
		min-inline-size: 0;
		min-block-size: 0;
		gap: var(--layout-section-gap);
	}

	.inspector {
		flex: none;
		inline-size: var(--layout-inspector-width);
		overflow: hidden;
	}

	/* Compact: the subsection list stacks above the content rather than beside
	   it — `layout-subnav-width` is 0 there, and a fixed-width column at zero
	   is not the same shape as a card in the page (ADR 0005). */
	@media (max-width: 767px) {
		.body {
			flex-direction: column;
		}
		.sections {
			inline-size: 100%;
		}
	}
</style>
