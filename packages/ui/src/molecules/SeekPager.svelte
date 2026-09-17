<script lang="ts">
	/**
	 * Mirrors `Seek pager` (`Hxmzc`). Paging for a list that seeks on its sort
	 * key rather than counting an offset.
	 *
	 * Not `Pagination` (`HlRZX`): that one states "41–60 of 1,284" and "3 / 65",
	 * and a seek knows neither number without a second count that would be
	 * stale before it rendered (ADR 0021). What replaces them is the span of
	 * dates on screen, which is what a person scanning backwards wants anyway.
	 *
	 * Older and Newer are links to a cursor rather than buttons that mutate in
	 * place, so Back works and a page is somewhere a member can return to.
	 */
	interface Props {
		/** The dates on screen: `17 September to 15 September 2026`, or one
		 *  date where the page covers a single day. */
		span: string;
		/** Where Older goes. Absent means there is nothing older. */
		olderHref?: string;
		/** Where Newer goes. Absent means this is the newest page. */
		newerHref?: string;
		/** How far back the ledger reaches, stated on the oldest page instead
		 *  of a control that would do nothing. */
		nothingOlder?: string;
	}

	let { span, olderHref, newerHref, nothingOlder }: Props = $props();
</script>

<nav class="seek-pager" aria-label="Pages of transactions">
	<span class="span">{span}</span>

	{#if !olderHref && nothingOlder}
		<span class="end">{nothingOlder}</span>
	{/if}

	<span class="paging">
		{#if newerHref}
			<a class="control" href={newerHref} rel="prev">Newer</a>
		{/if}
		{#if olderHref}
			<a class="control" href={olderHref} rel="next">Older</a>
		{/if}
	</span>
</nav>

<style>
	.seek-pager {
		display: flex;
		align-items: center;
		gap: 12px;
		inline-size: 100%;
		block-size: 56px;
		padding-inline: var(--density-cell-padding-x);
		background: var(--color-bg-surface);
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

	.paging {
		display: flex;
		align-items: center;
		gap: 8px;
		flex: 0 0 auto;
	}

	.control {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		block-size: 32px;
		padding-inline: 14px;
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-sm);
		background: var(--color-action-secondary);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 13px;
		text-decoration: none;
	}

	.control:hover {
		background: var(--color-action-secondary-hover);
	}

	.control:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}
</style>
