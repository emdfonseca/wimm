<script lang="ts">
	/**
	 * Mirrors `Account selector` (`vHJTa`), extended additively with a
	 * `Reading` line and a `Badge` slot — both disabled by default on the
	 * origin, so every instance that existed before renders unchanged.
	 *
	 * A balance is never shown without the time it was read. A stale reading
	 * keeps its figure and its original time and says it could not be updated:
	 * losing the previous number is the one thing that must not happen when a
	 * bank does not answer.
	 *
	 * What is rendered depends on the viewer's level. At `balance` the bank,
	 * the name and the amount appear and no identifier does — the server does
	 * not send one, and this does not invent one.
	 */
	interface Props {
		name: string;
		/** The bank. Always shown. */
		bank: string;
		/** Enough of the number to tell two accounts at one bank apart. Present
		 *  only at `details`; absent is absent, not a placeholder. */
		numberSuffix?: string;
		/** Already formatted, carrying its currency and its sign. */
		balance?: string;
		/** When the balance was read, as a phrase a person reads. */
		readAt?: string;
		/** The reading is the previous one: the bank could not be reached. */
		stale?: boolean;
		/** Access has run out; this account is no longer updating at all. */
		notUpdating?: boolean;
		/** True when the amount is below zero, so direction never rests on
		 *  colour alone — the figure itself carries the sign. */
		negative?: boolean;
		initials?: string;
	}

	let {
		name,
		bank,
		numberSuffix,
		balance,
		readAt,
		stale = false,
		notUpdating = false,
		negative = false,
		initials
	}: Props = $props();

	const mark = $derived(initials ?? bank.slice(0, 2).toUpperCase());
	const meta = $derived(numberSuffix ? `${bank} · •••• ${numberSuffix}` : bank);
</script>

<div class="account-row">
	<span class="mark" aria-hidden="true">{mark}</span>

	<span class="identity">
		<span class="name">{name}</span>
		<span class="meta">{meta}</span>
	</span>

	{#if notUpdating}
		<span class="badge stopped">Not updating</span>
	{:else if stale}
		<span class="badge">Could not update</span>
	{/if}

	{#if balance}
		<span class="figures">
			<span class="balance" class:negative>{balance}</span>
			{#if readAt}
				<span class="reading">
					{stale || notUpdating ? `Last read ${readAt}` : `Read ${readAt}`}
				</span>
			{/if}
		</span>
	{/if}
</div>

<style>
	.account-row {
		display: flex;
		align-items: center;
		gap: 12px;
		inline-size: 100%;
		block-size: 56px;
		padding-inline: 10px;
		border-radius: var(--radius-sm);
	}

	.mark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex: 0 0 auto;
		inline-size: 32px;
		block-size: 32px;
		/* Avatar origin `Z2oYAy`: a circle on bg-inverse with inverse text,
		   instanced at 32 in a row. It was a light square, which is why the
		   marks looked nothing like the frame. */
		border-radius: var(--radius-full);
		background: var(--color-bg-inverse);
		color: var(--color-text-inverse);
		font-family: var(--type-family-body);
		font-size: 12px;
		font-weight: 600;
	}

	.identity {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1 1 auto;
		min-inline-size: 0;
	}

	.name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 13px;
		font-weight: 500;
	}

	.meta {
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
	}

	.badge {
		flex: 0 0 auto;
		padding: 2px 8px;
		border-radius: var(--radius-pill);
		background: var(--color-feedback-warning-bg);
		color: var(--color-feedback-warning);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 500;
	}

	.stopped {
		background: var(--color-feedback-danger-bg);
		color: var(--color-feedback-danger);
	}

	.figures {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 0 0 auto;
		inline-size: 108px;
		text-align: end;
	}

	.balance {
		color: var(--color-amount-neutral);
		font-family: var(--type-family-mono);
		font-size: 13px;
	}

	/* The sign is what carries direction. The colour is additive, never the
	   only signal. */
	.negative {
		color: var(--color-amount-expense);
	}

	.reading {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}
</style>
