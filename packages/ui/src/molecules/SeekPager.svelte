<script lang="ts">
	/**
	 * Mirrors `Seek pager` (`Hxmzc`). States where a member is in a list that
	 * seeks on its sort key rather than counting an offset.
	 *
	 * Not `Pagination` (`HlRZX`): that one states "41–60 of 1,284" and "3 / 65",
	 * and a seek knows neither number without a second count that would be
	 * stale before it rendered (ADR 0021). What replaces them is the span of
	 * dates on screen, which is what a person scanning backwards wants anyway.
	 *
	 * Older and Newer moved to the scrubber, which already reaches every page
	 * this control could — a click on any dot is a page, exactly as Older and
	 * Newer were, and the ends of its track are Oldest and Newest. Carrying
	 * both was two controls answering the same question.
	 */
	interface Props {
		/** The dates on screen: `17 September to 15 September 2026`, or one
		 *  date where the page covers a single day. */
		span: string;
		/** Absent means there is nothing older to say so about. */
		nothingOlder?: string;
	}

	let { span, nothingOlder }: Props = $props();
</script>

<nav class="seek-pager" aria-label="Pages of transactions">
	<span class="span">{span}</span>

	{#if nothingOlder}
		<span class="end">{nothingOlder}</span>
	{/if}
</nav>

<style>
	.seek-pager {
		display: flex;
		align-items: center;
		gap: 12px;
		flex: 1 1 auto;
		min-inline-size: 0;
		block-size: 44px;
		padding-inline: var(--density-cell-padding-x);
	}

	.span {
		flex: 1 1 auto;
		min-inline-size: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
	}

	/* On the oldest page the screen says there is nothing older rather than
	   greying a control: a disabled control invites someone to work out how to
	   enable it. */
	.end {
		flex: 0 1 auto;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

</style>
