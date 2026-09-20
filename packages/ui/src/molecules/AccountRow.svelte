<script lang="ts">
	/**
	 * Mirrors `Account selector` (`vHJTa`), extended additively with a
	 * `Reading` line, a `Badge` slot, and the left-out state from `NZhzh` —
	 * all disabled by default on the origin, so every instance that existed
	 * before renders unchanged.
	 *
	 * A balance is never shown without the time it was read. A stale reading
	 * keeps its figure and its original time and says it could not be updated:
	 * losing the previous number is the one thing that must not happen when a
	 * bank does not answer.
	 *
	 * What is rendered depends on the viewer's level. At `balance` the bank,
	 * the name and the amount appear and no identifier does — the server does
	 * not send one, and this does not invent one.
	 *
	 * Left out (ADR 0022): the row dims and its figures never render, whatever
	 * is passed for them — a left-out account is never read, so there is
	 * nothing current to show — and a "Left out" badge replaces any other.
	 * This reaches only an owner; visibleAccountsQuery never returns a
	 * left-out account to anyone else, so no other viewer ever sees this row
	 * in this state at all.
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
		/** The account is left out of wimm. Reaches only an owner. */
		leftOut?: boolean;
		/** Makes the whole row one link and nothing else: a read-only row, as
		 *  Overview lists accounts. Absent, the row is plain and its owner
		 *  places controls beside it. */
		href?: string;
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
		initials,
		leftOut = false,
		href
	}: Props = $props();

	const mark = $derived(initials ?? bank.slice(0, 2).toUpperCase());
	const meta = $derived(numberSuffix ? `${bank} · •••• ${numberSuffix}` : bank);
</script>

<svelte:element
	this={href ? 'a' : 'div'}
	class="account-row"
	class:left-out={leftOut}
	class:link={!!href}
	{href}
>
	<span class="mark" aria-hidden="true">{mark}</span>

	<span class="identity">
		<span class="name">{name}</span>
		<span class="meta">{meta}</span>
	</span>

	{#if leftOut}
		<span class="badge left-out-badge">Left out</span>
	{:else if notUpdating}
		<span class="badge stopped">Not updating</span>
	{:else if stale}
		<span class="badge">Could not update</span>
	{/if}

	{#if balance && !leftOut}
		<span class="figures">
			<span class="balance" class:negative>{balance}</span>
			{#if readAt}
				<span class="reading">
					{stale || notUpdating ? `Last read ${readAt}` : `Read ${readAt}`}
				</span>
			{/if}
		</span>
	{/if}
</svelte:element>

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

	.link {
		color: inherit;
		text-decoration: none;
	}

	.link:hover {
		background: var(--color-bg-hover);
	}

	.link:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
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

	.left-out {
		background: var(--color-bg-subtle);
	}

	.left-out .name {
		color: var(--color-text-secondary);
	}

	.left-out-badge {
		background: var(--color-bg-elevated);
		color: var(--color-text-secondary);
	}

	.figures {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 0 0 auto;
		/* A floor, so short figures still line up down a list; "Read yesterday at
		   18:04" is wider than it and must not wrap beside an empty row. */
		min-inline-size: 108px;
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
		white-space: nowrap;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 11px;
	}
</style>
