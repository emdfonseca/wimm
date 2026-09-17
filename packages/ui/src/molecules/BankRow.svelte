<script lang="ts">
	/**
	 * Origin `hHYs5`. Height 56 to match the account row it sits beside,
	 * `radius-sm`, 10 of horizontal padding, no fill of its own at rest.
	 *
	 * One selectable bank in a list. It is a button rather than a link: picking
	 * a bank begins a hand-off out of the product, which is an action, and a
	 * link would offer to open it in a new tab where the return has nowhere to
	 * land.
	 */
	interface Props {
		name: string;
		/** Two letters standing in for a logo. The bank's own mark is not
		 *  always available, and a gap is worse than initials. */
		initials?: string;
		logoUrl?: string;
		disabled?: boolean;
		onselect?: () => void;
	}

	let { name, initials, logoUrl, disabled = false, onselect }: Props = $props();

	/** Falls back to the first letters of the bank's own name. */
	const mark = $derived(
		initials ??
			name
				.split(/\s+/)
				.slice(0, 2)
				.map((word) => word[0] ?? '')
				.join('')
				.toUpperCase()
	);
</script>

<button type="button" class="bank-row" {disabled} onclick={onselect}>
	{#if logoUrl}
		<img class="mark" src={logoUrl} alt="" />
	{:else}
		<span class="mark initials" aria-hidden="true">{mark}</span>
	{/if}
	<span class="name">{name}</span>
	<svg class="chevron" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
		<path
			d="m9 18 6-6-6-6"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
		/>
	</svg>
</button>

<style>
	.bank-row {
		display: flex;
		align-items: center;
		gap: 12px;
		inline-size: 100%;
		block-size: 56px;
		padding-inline: 10px;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		text-align: start;
		cursor: pointer;
		transition: background-color var(--motion-duration-fast) var(--motion-ease-standard);
	}

	/* The library has an Avatar origin but no Svelte component yet, so the mark
	   is inline here. It moves behind an Avatar import the moment that atom
	   exists, and the geometry below is the origin's. */
	.mark {
		flex: 0 0 auto;
		inline-size: 32px;
		block-size: 32px;
		border-radius: var(--radius-sm);
		object-fit: cover;
	}

	.initials {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		background: var(--color-bg-subtle);
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
		font-weight: 500;
	}

	.name {
		flex: 1 1 auto;
		min-inline-size: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 13px;
		font-weight: 500;
	}

	.chevron {
		flex: 0 0 auto;
		inline-size: 16px;
		block-size: 16px;
		color: var(--color-text-secondary);
	}

	.bank-row:hover:not(:disabled) {
		background: var(--color-bg-hover);
	}
	.bank-row:active:not(:disabled) {
		background: var(--color-bg-active);
	}

	/* Additive: the ring is drawn outside, so the row's own fill stays
	   underneath in every state. */
	.bank-row:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.bank-row:disabled {
		cursor: not-allowed;
	}
	.bank-row:disabled .name {
		color: var(--color-text-disabled);
	}

	@media (prefers-reduced-motion: reduce) {
		.bank-row {
			transition: none;
		}
	}
</style>
