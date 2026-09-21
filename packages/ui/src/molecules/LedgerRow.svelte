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
		/** Which of the member's accounts it came from. */
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
		/** Compact stacks the row into lines; wide lays it out in columns. */
		compact?: boolean;
		/** The statement text as the bank wrote it. Shown under the name only
		 *  where it differs from it: a name wimm derived has to be checkable
		 *  against the statement it came from. */
		banksLine?: string;
		/** Transactions files a row under a day header, so the row itself
		 *  carries no date in columns. Compact keeps it in the second line. */
		hideDate?: boolean;
		/** Overview's lists reserve no room for a marker, so the description
		 *  takes the width the slot would hold. A row that does carry a status
		 *  still shows it: this gives up the reserved column, not the fact. */
		hideStatus?: boolean;
		/** The one payment rule marked it: `Unusual`, or `Unusual income` when
		 *  the row is not `negative`. Sits in the status slot, which a booked row
		 *  never otherwise uses. */
		unusual?: boolean;
		/** The row is one half of a movement between two accounts the member
		 *  owns: `Between your accounts`. Shares the status slot and wins it
		 *  from `unusual`, because a transfer is never also unusual. `Not
		 *  settled` still wins over both. */
		transfer?: boolean;
		/** A second line under the name, in the bank's line's place and never
		 *  beside one: `usually about €60`. */
		note?: string;
		/** Makes the whole row one link. */
		href?: string;
	}

	let {
		description,
		account,
		amount,
		date,
		negative = false,
		unsettled = false,
		initials,
		compact = false,
		banksLine,
		hideDate = false,
		hideStatus = false,
		unusual = false,
		transfer = false,
		note,
		href
	}: Props = $props();

	const mark = $derived(initials ?? (description.slice(0, 2).toUpperCase() || '—'));
	const line = $derived(note ?? (banksLine && banksLine !== description ? banksLine : undefined));
	const unusualLabel = $derived(negative ? 'Unusual' : 'Unusual income');
	// One row shows at most one status: Not settled, then Between your
	// accounts, then Unusual. The load never sends two, and the order is here
	// so a row that carries both cannot render both.
	const status = $derived(
		unsettled ? 'Not settled' : transfer ? 'Between your accounts' : unusual ? unusualLabel : undefined
	);
</script>

<svelte:element
	this={href ? 'a' : 'div'}
	class="ledger-row"
	class:compact
	class:with-line={compact && !!line}
	class:link={!!href}
	{href}
>
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
				<span class="when">{date}</span>
				{#if unsettled}
					<span class="badge">Not settled</span>
				{:else if status}
					<span class="tag">{status}</span>
				{/if}
			</span>
			{#if line}
				<span class="banks-line" title={line}>{line}</span>
			{/if}
		</span>
	{:else}
		<span class="named">
			<span class="description">{description}</span>
			{#if line}
				<span class="banks-line" title={line}>{line}</span>
			{/if}
			<!-- A list with no marker column has no room for a long label
			     beside the name, and squeezing it there costs the name its
			     words. It goes under the name instead, where it is whole. -->
			{#if hideStatus && status}
				<span class="second-line">
					{#if unsettled}
						<span class="badge">Not settled</span>
					{:else}
						<span class="tag">{status}</span>
					{/if}
				</span>
			{/if}
		</span>
		<span class="account">{account}</span>
		{#if !hideStatus}
			<span class="status">
				{#if unsettled}
					<span class="badge">Not settled</span>
				{:else if status}
					<span class="tag">{status}</span>
				{/if}
			</span>
		{/if}
		{#if !hideDate}
			<span class="date">{date}</span>
		{/if}
		<span class="amount" class:negative>{amount}</span>
	{/if}
</svelte:element>

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

	.link {
		color: inherit;
		text-decoration: none;
	}

	.link:hover {
		background: var(--color-bg-hover);
	}

	.link:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: calc(-1 * var(--focus-ring-width));
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

	.compact.with-line {
		block-size: 80px;
	}

	.compact .mark {
		inline-size: 32px;
		block-size: 32px;
		font-size: 12px;
	}

	/* Sized from its words unless the caller says otherwise. In a table the
	   caller sets --ledger-name-basis to 0 so the name takes what the columns
	   leave: sized from a long bank's line it pushes the account out of its
	   column, and the rows of a Medium table stop lining up. */
	.named {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1 1 var(--ledger-name-basis, auto);
		min-inline-size: 0;
	}

	.named .description {
		flex: 0 1 auto;
	}

	.banks-line {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-placeholder);
		font-family: var(--type-family-body);
		font-size: 11px;
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

	/* In a table the account is a column and holds its width so rows line up.
	   A caller whose rows are too narrow for that column, where it squeezes the
	   name to "Pingo D…" beside the account's own empty space, sets
	   --ledger-account-basis to auto and the account is as wide as its words. */
	.account {
		flex: 0 1 var(--ledger-account-basis, 220px);
		min-inline-size: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: 12px;
	}

	/* A column, not a tag. It holds its width for the reason the account
	   above it does: sized from its own words, the longest label indents the
	   account of the one row that carries it and the table stops lining up.
	   The basis is the widest label, `Between your accounts`. */
	.status {
		display: inline-flex;
		align-items: center;
		flex: 0 0 var(--ledger-status-basis, 148px);
		block-size: 22px;
	}

	/* Under the name, in the bank's line's place: its own line, so the label
	   is whole and the name keeps its words. */
	.second-line {
		display: inline-flex;
		align-items: center;
		align-self: start;
		block-size: 22px;
	}

	.date {
		flex: 0 0 auto;
		min-inline-size: 72px;
		white-space: nowrap;
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

	.tag {
		flex: 0 0 auto;
		padding: 2px 8px;
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-pill);
		color: var(--color-text-secondary);
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

	.when {
		flex: 0 0 auto;
		white-space: nowrap;
	}
</style>
