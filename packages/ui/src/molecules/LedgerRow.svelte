<script lang="ts">
	/**
	 * Mirrors `Ledger row` (`VE2z9`). One transaction as the ledger shows it.
	 *
	 * Not `Transaction row` (`PviJn`): that one carries a category column this
	 * change does not have, and carries neither the account nor a settled
	 * marker, which it needs. The word here is Description rather than
	 * Merchant, because a salary and a transfer between accounts have no
	 * merchant.
	 *
	 * Direction is carried by the sign on the amount. The colour is additive
	 * and never the only signal (ADR 0006).
	 */
	interface Props {
		/** What it was. Empty where the bank named nobody — never invented. */
		description: string;
		/** Which of the member's accounts it came from, and its bank. */
		account: string;
		/** Already formatted, carrying its currency and its sign. */
		amount: string;
		/** The day, short: `17 Sep`. */
		date: string;
		/** Below zero, so the row can tint what the sign already says. */
		negative?: boolean;
		/** The bank has not settled it. Money that has moved, as far as the
		 *  member is concerned, so it is listed and marked rather than hidden. */
		unsettled?: boolean;
		/** Two letters for the mark. Derived from the description otherwise. */
		initials?: string;
		/** Compact stacks the row into two lines; wide lays it out in columns. */
		compact?: boolean;
	}

	let {
		description,
		account,
		amount,
		date,
		negative = false,
		unsettled = false,
		initials,
		compact = false
	}: Props = $props();

	const mark = $derived(initials ?? (description.slice(0, 2).toUpperCase() || '—'));
</script>

<div class="ledger-row" class:compact>
	<span class="mark" aria-hidden="true">{mark}</span>

	{#if compact}
		<span class="stack">
			<span class="line">
				<span class="description">{description}</span>
				<span class="amount" class:negative>{amount}</span>
			</span>
			<span class="line meta">
				<span class="account">{account}</span>
				<span aria-hidden="true">·</span>
				<span>{date}</span>
				{#if unsettled}
					<span class="badge">Not settled</span>
				{/if}
			</span>
		</span>
	{:else}
		<span class="description">{description}</span>
		<span class="account">{account}</span>
		<span class="status">
			{#if unsettled}
				<span class="badge">Not settled</span>
			{/if}
		</span>
		<span class="date">{date}</span>
		<span class="amount" class:negative>{amount}</span>
	{/if}
</div>

<style>
	.ledger-row {
		display: flex;
		align-items: center;
		gap: 12px;
		inline-size: 100%;
		block-size: var(--density-row-height);
		padding-inline: var(--density-cell-padding-x);
		background: var(--color-bg-elevated);
	}

	.compact {
		gap: 11px;
		block-size: 64px;
	}

	.mark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex: 0 0 auto;
		inline-size: 28px;
		block-size: 28px;
		border-radius: var(--radius-full);
		background: var(--color-bg-inverse);
		color: var(--color-text-inverse);
		font-family: var(--type-family-body);
		font-size: 11px;
		font-weight: 600;
	}

	.compact .mark {
		inline-size: 32px;
		block-size: 32px;
		font-size: 12px;
	}

	.description {
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

	.account {
		flex: 0 0 auto;
		inline-size: 220px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
	}

	.status {
		display: inline-flex;
		align-items: center;
		flex: 0 0 auto;
		inline-size: 92px;
		block-size: 22px;
	}

	.date {
		flex: 0 0 auto;
		inline-size: 52px;
		text-align: end;
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
	}

	.amount {
		flex: 0 0 auto;
		inline-size: 112px;
		text-align: end;
		color: var(--color-amount-income);
		font-family: var(--type-family-mono);
		font-size: 13px;
	}

	/* The sign is what carries direction; the colour is additive. */
	.negative {
		color: var(--color-amount-expense);
	}

	.badge {
		flex: 0 0 auto;
		padding: 2px 8px;
		border-radius: var(--radius-pill);
		background: var(--color-feedback-warning-bg);
		color: var(--color-feedback-warning);
		font-family: var(--type-family-body);
		font-size: 11px;
		font-weight: 500;
		white-space: nowrap;
	}

	/* Compact stacks two lines beside the mark, which is what lets a 390 px
	   screen carry the same five facts the wide columns do. */
	.stack {
		display: flex;
		flex-direction: column;
		gap: 3px;
		flex: 1 1 auto;
		min-inline-size: 0;
	}

	.line {
		display: flex;
		align-items: center;
		gap: 10px;
		inline-size: 100%;
	}

	.meta {
		gap: 6px;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}

	.compact .account {
		inline-size: auto;
		flex: 0 1 auto;
		font-size: 11px;
	}

	.compact .amount {
		inline-size: 96px;
	}
</style>
