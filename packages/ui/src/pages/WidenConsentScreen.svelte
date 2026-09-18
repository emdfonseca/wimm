<script lang="ts" module>
	/**
	 * J08.A / 01 · What wimm will see · Widening.
	 *
	 * The screen a member reaches from the notice on Transactions when a bank
	 * was connected before wimm could read them. It is not the first-connection
	 * explainer: the member has already granted something, so the question is
	 * what is being added rather than what is being given.
	 *
	 * The connection is live and narrow, which is neither working nor expired.
	 * Nothing here calls it broken.
	 */
</script>

<script lang="ts">
	import Button from '../atoms/Button.svelte';
	import Page from '../templates/Page.svelte';

	interface Props {
		bankName: string;
		/** Already formatted as a person reads it, e.g. "17 December 2026". */
		accessEndsOn: string;
		oncontinue?: () => void;
		/** Where "Not now" goes. Declining leaves the connection exactly as it
		 *  was, still working for balances. */
		cancelHref?: string;
	}

	let { bankName, accessEndsOn, oncontinue, cancelHref = '/transactions' }: Props = $props();
</script>

<Page title="Include {bankName}'s transactions" focusHeading>
	<div class="prose">
		<p class="lede">
			You already let wimm read this bank's account names and balances. Confirming once more adds
			its transactions, so wimm can show what happened on the accounts you own.
		</p>

		<ul class="scope">
			<li>Account names and balances, as before</li>
			<li>Transactions on the accounts you own</li>
			<li class="never">Never your login, your password or your card number</li>
		</ul>

		<p class="helper">
			{bankName}'s access will run out on {accessEndsOn}. That date is {bankName}'s own limit, not a
			wimm setting.
		</p>

		<div class="actions">
			<Button onclick={oncontinue}>Continue to {bankName}</Button>
			<a class="not-now" href={cancelHref}>Not now</a>
		</div>
	</div>
</Page>

<style>
	/* Prose, so its measure is capped (ADR 0005). */
	.prose {
		display: flex;
		flex-direction: column;
		gap: 16px;
		max-inline-size: var(--layout-content-max);
		font-family: var(--type-family-body);
	}

	.lede {
		margin: 0;
		max-inline-size: 564px;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-md);
	}

	.scope {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding-inline-start: 20px;
		color: var(--color-text-primary);
		font-size: var(--type-size-body-md);
	}

	/* What is not granted reads as secondary, because it is the reassurance
	   rather than the offer. */
	.never {
		color: var(--color-text-secondary);
	}

	.helper {
		margin: 0;
		max-inline-size: 564px;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.not-now {
		color: var(--color-accent);
		font-size: var(--type-size-body-md);
		font-weight: 600;
		text-decoration: none;
	}

	.not-now:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}
</style>
