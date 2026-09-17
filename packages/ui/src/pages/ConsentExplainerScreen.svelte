<script lang="ts" module>
	/**
	 * J03.A / 02 · What wimm will see.
	 *
	 * The last screen before the member leaves for their bank. Its whole job is
	 * to say what is about to be granted and for how long, in the member's own
	 * terms, before they can no longer change it here.
	 *
	 * The date is the bank's, not a wimm policy — and the difference matters
	 * more than it looks: consent runs 90 days at some banks and 1 day at
	 * others, so a member who is not shown the real date will read a daily
	 * prompt as a defect.
	 */
</script>

<script lang="ts">
	import Button from '../atoms/Button.svelte';

	interface Props {
		bankName: string;
		/** Already formatted as a person reads it, e.g. "18 September 2026". */
		accessEndsOn: string;
		/** Set when that date is soon enough to be worth naming plainly. */
		shortLived?: boolean;
		oncontinue?: () => void;
		oncancel?: () => void;
	}

	let { bankName, accessEndsOn, shortLived = false, oncontinue, oncancel }: Props = $props();
</script>

<main class="screen">
	<h1 tabindex="-1">What {bankName} will share with wimm</h1>

	<ul class="scope">
		<li><strong>Your accounts and their balances.</strong> Names, numbers and amounts.</li>
		<li>
			<strong>Nothing else.</strong> Not your transactions, and nothing that could move money. wimm
			cannot make a payment.
		</li>
		<li>
			<strong>Not your banking password.</strong> You confirm at {bankName}. wimm never sees it.
		</li>
	</ul>

	<p class="until">
		Access ends on <strong>{accessEndsOn}</strong>, which is {bankName}'s own limit rather than a
		wimm setting.
		{#if shortLived}
			That is soon: you will be asked to confirm again, and wimm will tell you when.
		{:else}
			After that you confirm again to carry on.
		{/if}
	</p>

	<div class="actions">
		<Button onclick={oncontinue}>Continue to {bankName}</Button>
		<Button variant="secondary" onclick={oncancel}>Cancel</Button>
	</div>
</main>

<style>
	.screen {
		/* Fills the slot and packs content to the top, as the frame's Page does
		   with a trailing spacer. A screen that does not fill gets centred by
		   the slot instead, which is what left these floating mid-page. */
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 16px;
		inline-size: 100%;
		max-inline-size: var(--layout-content-max); /* prose, so its measure is capped (ADR 0005) */
	}

	h1 {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-display);
		font-size: var(--type-size-page-title);
		font-weight: 600;
	}
	h1:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.scope {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding-inline-start: 20px;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.scope strong,
	.until strong {
		color: var(--color-text-primary);
		font-weight: 600;
	}

	.until {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.actions {
		display: flex;
		gap: 12px;
		align-items: center;
	}

	@media (max-width: 599px) {
		.actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
