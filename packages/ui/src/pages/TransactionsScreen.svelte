<script lang="ts" module>
	/**
	 * J07 · See where the money went. The ledger, and every state it can be in.
	 *
	 * Everything here is **that member's own view**: the transactions of the
	 * accounts they own, and no others. A member granted *balance* or *details*
	 * on an account they do not own sees none of its transactions and is not
	 * told how many there are — seeing a balance and seeing what was spent are
	 * different sentences, and only the first was agreed to (ADR 0021).
	 */
	export interface LedgerEntry {
		id: string;
		/** The day this row is filed under, already grouped by the caller. */
		description: string;
		/** The account and its bank: `Current account · Monzo`. */
		account: string;
		amount: string;
		/** Short: `17 Sep`. */
		date: string;
		negative?: boolean;
		unsettled?: boolean;
		initials?: string;
		/** The statement text as the bank wrote it. Shown under `description`
		 *  only where it differs from it. */
		banksLine?: string;
		/** The one payment rule marked it. Labelled `Unusual`, or `Unusual
		 *  income` where it came in. */
		unusual?: boolean;
		/** One half of a movement between two accounts the member owns:
		 *  `Between your accounts`, in the status slot. */
		transfer?: boolean;
	}

	/** One day's transactions under the heading that names it. */
	export interface LedgerDay {
		/** The day itself, as `YYYY-MM-DD`. The heading's words are this
		 *  screen's, not the caller's, so "Today" cannot mean two things in two
		 *  places. A day straddling a page boundary is named again at the top of
		 *  the next page — the same heading, not a new one. */
		date: string;
		entries: LedgerEntry[];
	}

	function localIso(d: Date): string {
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
	}

	/** The words above a day's transactions, which is where a row's date
	 *  lives: `Today, 17 September`, `Yesterday, 16 September`, or an older
	 *  day with its year. `today` is `YYYY-MM-DD`. */
	export function dayLabel(date: string, today = localIso(new Date())): string {
		const [year, month, day] = date.split('-').map(Number);
		const [ty, tm, td] = today.split('-').map(Number);
		if (!year || !month || !day || !ty || !tm || !td) return date;

		const long = (options: Intl.DateTimeFormatOptions) =>
			new Date(Date.UTC(year, month - 1, day)).toLocaleDateString('en-GB', {
				...options,
				timeZone: 'UTC'
			});

		if (date === today) return `Today, ${long({ day: 'numeric', month: 'long' })}`;

		const yesterday = new Date(Date.UTC(ty, tm - 1, td - 1)).toISOString().slice(0, 10);
		if (date === yesterday) return `Yesterday, ${long({ day: 'numeric', month: 'long' })}`;

		return long({ day: 'numeric', month: 'long', year: 'numeric' });
	}

	/** What a filtered list adds up to in one currency, already written with
	 *  its sign: `+€2,180.00`, `−€1,204.55`. */
	export interface LedgerFigures {
		moneyIn: string;
		moneyOut: string;
	}

	/** A bank connected before wimm could read transactions. */
	export interface NarrowBank {
		connectionId: string;
		bankName: string;
		/** Where widening starts. */
		widenHref: string;
	}

	/** A bank that is not contributing to the list, and why. */
	export interface LedgerProblem {
		connectionId: string;
		bankName: string;
		kind:
			'unreachable' | 'rate-limited' | 'access-ended' | 'disconnected' | 'first-read' | 'widened';
		/** Only for rate-limited, and only when the bank said. */
		retryAfter?: string;
		/** Only for unreachable: when that bank's rows were last read. */
		lastReadAt?: string;
		/** Where restoring starts, for access-ended. */
		restoreHref?: string;
		/** Where asking again goes, for unreachable. */
		retryHref?: string;
	}
</script>

<script lang="ts">
	import Button from '../atoms/Button.svelte';
	import Icon from '../atoms/Icon.svelte';
	import EmptyState from '../molecules/EmptyState.svelte';
	import ErrorNotice from '../molecules/ErrorNotice.svelte';
	import InfoNotice from '../molecules/InfoNotice.svelte';
	import LedgerRow from '../molecules/LedgerRow.svelte';
	import SeekPager from '../molecules/SeekPager.svelte';
	import PageScrubber, { type ScrubberPage } from '../molecules/PageScrubber.svelte';
	import LedgerFilters, {
		filtersInForce,
		type FilterOption,
		type LedgerFilterValues
	} from '../molecules/LedgerFilters.svelte';
	import Page from '../templates/Page.svelte';

	interface Props {
		days?: LedgerDay[];
		/** How many transactions this member may see. A count of what they may
		 *  see, not a page count: it survives paging unchanged. */
		count?: number;
		/** `Updated at 09:14. Reaching back to 4 June 2026.` Absent before the
		 *  first read: wimm states what it reached rather than promising a
		 *  period no bank commits to in advance. */
		freshness?: string;
		/** The dates on this page. */
		span?: string;
		/** The oldest page has been reached, so the screen says there is nothing
		 *  older rather than offering a control that would do nothing. */
		atOldest?: boolean;

		/** Every real page the ledger holds for this scope, newest first, each
		 *  carrying where choosing it goes — never a calendar month standing
		 *  in for one (ADR 0021, extended). */
		pages?: ScrubberPage[];
		/** Which of those pages this one is, by its key, or null on a page
		 *  reached some other way (paging one at a time). */
		currentPage?: string | null;
		/** Where jumping to the newest page goes. Absent on the newest page. */
		newestHref?: string;
		/** Where jumping to the oldest page goes. Absent on the oldest page. */
		oldestHref?: string;

		/** The filters in force. The account select names the account the list
		 *  is narrowed to; a filtered list is the same screen, not another. */
		filters?: LedgerFilterValues;
		/** Every account the member can narrow to, as `Savings · Monzo`. */
		accounts?: FilterOption[];
		/** The months holding a match, newest first, as `August 2026`. */
		months?: FilterOption[];
		/** Where Clear filters and Show all accounts go. */
		clearHref?: string;
		/** What the filtered list adds up to, one entry per currency. Shown only
		 *  while a filter is in force and rows match. */
		totals?: LedgerFigures[];
		/** Transfers between the member's accounts the totals leave out. */
		transfersLeftOut?: number;
		/** Matching rows the totals do not count because they are not settled. */
		notSettled?: number;
		/** A filter changed; `live` when a pause in typing applied a search.
		 *  Absent, the filter bar submits as a plain GET. */
		onfilter?: (next: LedgerFilterValues, how: { live: boolean }) => void;

		narrow?: NarrowBank[];
		problems?: LedgerProblem[];

		/** No bank connected at all. */
		noBank?: boolean;
		/** This member owns no account, whatever they were granted on others. */
		ownsNothing?: boolean;
		/** Where connecting a first bank starts. */
		connectHref?: string;

		/** The date the screen calls Today, as `YYYY-MM-DD`. Left unset it is the
		 *  device's. Fixed by a story so its words do not move with the calendar. */
		today?: string;

		refreshing?: boolean;
		onrefresh?: () => void;
		/** Compact stacks each row into two lines. Left unset, the screen reads
		 *  its own regime — nothing above it knows about layout breakpoints,
		 *  and this is the one screen whose layout genuinely differs in
		 *  structure rather than only in CSS. */
		compact?: boolean;
	}

	let {
		days = [],
		count = 0,
		freshness,
		span,
		atOldest = false,
		pages = [],
		currentPage = null,
		newestHref,
		oldestHref,
		filters = { account: '', q: '', month: '', direction: '' },
		accounts = [],
		months = [],
		clearHref = '/transactions',
		totals = [],
		transfersLeftOut = 0,
		notSettled = 0,
		onfilter,
		narrow = [],
		problems = [],
		noBank = false,
		ownsNothing = false,
		connectHref = '/connect',
		today,
		refreshing = false,
		onrefresh,
		compact: compactProp
	}: Props = $props();

	/** `768px` is the declared Compact/Medium boundary (ADR 0002). Read
	 *  directly rather than passed down, because every caller would otherwise
	 *  need to know a layout breakpoint that is this screen's concern alone. */
	let autoCompact = $state(false);

	$effect(() => {
		if (typeof window === 'undefined' || !window.matchMedia) return;
		const query = window.matchMedia('(max-width: 767px)');
		autoCompact = query.matches;
		const onchange = (e: MediaQueryListEvent) => (autoCompact = e.matches);
		query.addEventListener('change', onchange);
		return () => query.removeEventListener('change', onchange);
	});

	const compact = $derived(compactProp ?? autoCompact);

	const hasRows = $derived(days.some((day) => day.entries.length > 0));
	const inForce = $derived(filtersInForce(filters));
	/** Only the account is in force: an empty list then says that account has
	 *  nothing, which is a different sentence from nothing matching. */
	const accountOnly = $derived(
		filters.account !== '' && filters.q === '' && filters.month === '' && filters.direction === ''
	);
	const accountLabel = $derived(
		accounts.find((a) => a.value === filters.account)?.label ?? 'This account'
	);

	/** The reasons for having nothing differ and must not be collapsed into one
	 *  empty list: each says which thing the member can do about it. */
	const emptyReason = $derived.by(() => {
		if (noBank) return 'no-bank' as const;
		if (ownsNothing) return 'owns-nothing' as const;
		// Rows are held behind the filters, so nothing here is about a bank.
		if (inForce) return 'no-match' as const;
		if (narrow.length > 0) return 'narrow' as const;
		// A sync running with nothing stored is the first read. The member is
		// told the list fills as the bank answers rather than being shown an
		// empty list that looks settled.
		if (refreshing || problems.some((p) => p.kind === 'first-read')) return 'first-read' as const;
		return 'nothing-read' as const;
	});

	/** The note under the totals is there for checking the figures, not for
	 *  reading every time: it opens on asking. */
	let noteOpen = $state(false);

	/** What the totals do not count, said so the sum can be checked against
	 *  the rows. Settled money only, and a transfer between two of the
	 *  member's accounts moves nothing, as on Overview. */
	const totalsNote = $derived(
		[
			transfersLeftOut > 0
				? `Leaves out ${transfersLeftOut} ${transfersLeftOut === 1 ? 'transfer' : 'transfers'} between your accounts.`
				: '',
			notSettled > 0
				? `Does not count ${notSettled} ${notSettled === 1 ? 'payment' : 'payments'} not yet settled.`
				: ''
		]
			.filter(Boolean)
			.join(' ')
	);

	/** Nothing to filter where there is no bank or no account of their own. */
	const showFilters = $derived(!noBank && !ownsNothing && (hasRows || inForce));

	const noMatchTitle = $derived(
		accountOnly ? `${accountLabel} has no transactions` : 'Nothing matches these filters'
	);

	/** The bank whose transactions are not being read yet, for the empty state
	 *  that names it. */
	const narrowBank = $derived(narrow[0]?.bankName ?? '');

	// The account alone narrows what the bank sent, so its oldest row is still
	// as far back as the bank would go. Any other filter ends before that.
	const nothingOlder = $derived.by(() => {
		if (!atOldest) return undefined;
		return inForce && !accountOnly
			? 'Nothing older matches these filters.'
			: 'Nothing older. This is as far back as the bank would go.';
	});

	let listHeading = $state<HTMLElement | null>(null);

	/**
	 * A member who asked to refresh gets no visual cue if nothing changed, so
	 * what happened is announced rather than only drawn. The canvas cannot draw
	 * a live region; it is a contract it states.
	 */
	const refreshOutcome = $derived.by(() => {
		if (refreshing) return 'Bringing transactions up to date';
		if (problems.length === 0) return '';

		const limited = problems.find((p) => p.kind === 'rate-limited');
		if (limited) {
			return limited.retryAfter
				? `${limited.bankName} has been asked too often. ${limited.bankName} will accept another request after ${limited.retryAfter}. What is below was already read, and it has not changed.`
				: `${limited.bankName} has been asked too often. What is below was already read, and it has not changed.`;
		}

		const unreachable = problems.filter((p) => p.kind === 'unreachable').map((p) => p.bankName);
		if (unreachable.length > 0) {
			return `${unreachable.join(', ')} did not answer. Everything else is up to date.`;
		}
		return '';
	});

	/**
	 * A member whose focus stayed on the filter they just changed sees nothing
	 * move under it, so the outcome is announced, ahead of any bank that did
	 * not answer. While a refresh runs, that is all there is to say.
	 */
	const status = $derived.by(() => {
		if (refreshing || !inForce) return refreshOutcome;
		const outcome = hasRows ? `${count} transactions` : noMatchTitle;
		return [outcome, refreshOutcome].filter(Boolean).join('. ');
	});
</script>

{#snippet problemNotice(problem: LedgerProblem)}
	{#if problem.kind === 'rate-limited'}
		<ErrorNotice title="{problem.bankName} has been asked too often" live="polite">
			{problem.bankName} will accept another request after {problem.retryAfter}. What is below was
			already read, and it has not changed.
		</ErrorNotice>
	{:else if problem.kind === 'unreachable'}
		<ErrorNotice title="{problem.bankName} did not answer" live="polite">
			Everything else is up to date. {problem.bankName}'s transactions are the ones last read at {problem.lastReadAt}.
			<a class="inline-action" href={problem.retryHref ?? '/transactions?refresh=1'}>Try again</a>
		</ErrorNotice>
	{:else if problem.kind === 'access-ended'}
		<ErrorNotice title="{problem.bankName} has stopped sending transactions" live="polite">
			The access you gave {problem.bankName} has run out. What is below stays here, and nothing new arrives
			until you confirm again at the bank.
			<a class="inline-action" href={problem.restoreHref ?? '/'}>Restore {problem.bankName}</a>
		</ErrorNotice>
	{:else if problem.kind === 'disconnected'}
		<InfoNotice title="{problem.bankName} is no longer connected">
			Its accounts have left Overview and its balances are gone. The transactions already read are
			still here, and they stay whether or not you connect {problem.bankName} again.
		</InfoNotice>
	{:else if problem.kind === 'first-read'}
		<InfoNotice title="Reading your transactions">
			{problem.bankName} is sending them now. This is the first time, so it can take a moment. How far
			back they reach is whatever {problem.bankName} gives.
		</InfoNotice>
	{:else if problem.kind === 'widened'}
		<InfoNotice title="{problem.bankName} is sending transactions now">
			This is the first read, so it goes back as far as {problem.bankName} will give. Nothing about who
			owns these accounts has changed.
		</InfoNotice>
	{/if}
{/snippet}

<Page title="Transactions">
	{#snippet action()}
		<!-- How current the list is sits with the action that changes it,
		     leaving the space above the list to the list. -->
		<div class="header-action">
			{#if freshness && !compact}
				<p class="updated">{freshness}</p>
			{/if}
			{#if hasRows || inForce || problems.length > 0 || refreshing}
				<Button onclick={onrefresh} disabled={refreshing}>
					{refreshing ? 'Refreshing…' : 'Refresh'}
				</Button>
			{/if}
		</div>
	{/snippet}

	<div class="screen" class:compact>
		<!-- A phone's header has room for the title and Refresh only. -->
		{#if compact && freshness}
			<p class="helper">{freshness}</p>
		{/if}

		<!-- A bank that cannot be read for transactions is an alert on that bank,
	     never a page banner: a household with one narrow bank and two current
	     ones would otherwise be warned about the whole product, daily. -->
		{#each narrow as bank (bank.connectionId)}
			<InfoNotice title="{bank.bankName} is not sending transactions yet">
				This bank was connected before wimm could read transactions. Confirm once more at {bank.bankName}
				and they will start arriving. Your balances are unaffected.
				<a class="inline-action" href={bank.widenHref}>Widen at {bank.bankName}</a>
			</InfoNotice>
		{/each}

		{#each problems as problem (problem.connectionId)}
			{@render problemNotice(problem)}
		{/each}

		<p class="sr-only" role="status" aria-live="polite">{status}</p>

		<!-- A filter is not containment: the account select names the account,
		     there is no breadcrumb, and clearing widens the same list. -->
		{#if showFilters}
			<LedgerFilters
				{filters}
				{accounts}
				{months}
				{clearHref}
				{compact}
				{onfilter}
				offerClear={hasRows || accountOnly}
				onclear={() => listHeading?.focus()}
			/>
		{/if}

		{#if !hasRows}
			{#if emptyReason === 'no-match'}
				<EmptyState title={noMatchTitle} elevated>
					{accountOnly
						? 'The bank has sent none for this account.'
						: 'Change a filter above, or clear them to see every transaction.'}
					{#snippet action()}
						<a class="inline-action" href={clearHref} onclick={() => listHeading?.focus()}>
							{accountOnly ? 'Show all accounts' : 'Clear filters'}
						</a>
					{/snippet}
				</EmptyState>
			{:else if emptyReason === 'no-bank'}
				<EmptyState title="No transactions yet" elevated>
					Connect a bank and wimm will read what happens on your accounts. Nobody else in the
					household sees any of it until you say so.
					{#snippet action()}
						<a class="inline-action" href={connectHref}>Connect a bank</a>
					{/snippet}
				</EmptyState>
			{:else if emptyReason === 'owns-nothing'}
				<EmptyState title="You do not own any accounts" elevated>
					Transactions are shown for accounts that are yours. Ask whoever connected the bank to make
					you an owner of one.
				</EmptyState>
			{:else if emptyReason === 'narrow'}
				<EmptyState title="Nothing read from {narrowBank}" elevated>
					There is nothing to show until the bank starts sending transactions.
				</EmptyState>
			{:else if emptyReason === 'first-read'}
				<EmptyState title="Nothing read yet" elevated
					>The list fills as the bank answers.</EmptyState
				>
			{:else}
				<EmptyState title="Nothing read yet" elevated
					>The list fills as the bank answers.</EmptyState
				>
			{/if}
		{:else}
			<section class="ledger" aria-labelledby="ledger-heading">
				<div class="toolbar">
					<h2 id="ledger-heading" class="count" tabindex="-1" bind:this={listHeading}>
						{count} transactions
					</h2>
					<!-- What the filtered list adds up to, on the count's own line:
					     one pair of figures per currency, never summed across them. -->
					{#if inForce && totals.length > 0}
						<div class="totals">
							{#each totals as total, i (i)}
								<p class="figures">
									{#if filters.direction !== 'out'}
										<span>In <span class="amount income">{total.moneyIn}</span></span>
									{/if}
									{#if filters.direction !== 'in'}
										<span>Out <span class="amount expense">{total.moneyOut}</span></span>
									{/if}
								</p>
							{/each}
							{#if totalsNote}
								<button
									type="button"
									class="why"
									aria-label="What these figures leave out"
									aria-expanded={noteOpen}
									aria-controls="totals-note"
									onclick={() => (noteOpen = !noteOpen)}
								>
									<Icon name="info" size={14} />
								</button>
							{/if}
						</div>
					{/if}
				</div>
				{#if totalsNote && noteOpen}
					<p id="totals-note" class="note">{totalsNote}</p>
				{/if}

				{#if !compact}
					<div class="columns" aria-hidden="true">
						<span class="col-gap"></span>
						<span class="col-description">Description</span>
						<span class="col-account">Account</span>
						<span class="col-status"></span>
						<span class="col-amount">Amount</span>
					</div>
				{/if}

				<!-- Only this scrolls: the ledger's own height never exceeds the
			     screen, and the toolbar, columns and pager stay put around
			     whatever part of the list is on screen. -->
				<!-- The list scrolls, so a keyboard has to be able to reach it. -->
				<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
				<div class="rows" role="region" aria-label="Transactions" tabindex="0">
					{#each days as day (day.date + day.entries[0]?.id)}
						<h3 class="day">{dayLabel(day.date, today)}</h3>
						{#each day.entries as entry (entry.id)}
							<LedgerRow
								description={entry.description}
								account={entry.account}
								amount={entry.amount}
								date={entry.date}
								negative={entry.negative}
								unsettled={entry.unsettled}
								initials={entry.initials}
								banksLine={entry.banksLine}
								unusual={entry.unusual}
								transfer={entry.transfer}
								hideDate={!compact}
								{compact}
							/>
						{/each}
					{/each}
				</div>

				<!-- One row, not two: the span says where the member is and the
			     scrubber is how they move, and neither needs a whole row to
			     itself. -->
				{#if span || pages.length > 0}
					<div class="pager" class:compact>
						{#if span}
							<SeekPager {span} {nothingOlder} />
						{/if}
						{#if pages.length > 0}
							<PageScrubber {pages} current={currentPage} {newestHref} {oldestHref} />
						{/if}
					</div>
				{/if}
			</section>
		{/if}
	</div>
</Page>

<style>
	/* Fills the slot and packs content to the top, as the frame's Page does.
	   A screen that does not fill gets centred by the slot instead. */
	.screen {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-block-size: 0;
		gap: var(--space-4);
		inline-size: 100%;
		font-family: var(--type-family-body);
	}

	.compact {
		gap: var(--space-4);
	}

	.header-action {
		display: flex;
		align-items: center;
		gap: var(--space-4);
	}

	.updated {
		margin: 0;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
		text-align: end;
	}

	.helper {
		margin: 0;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
	}

	.inline-action {
		color: var(--color-accent);
		font-size: var(--type-size-body-sm);
		font-weight: 600;
		text-decoration: none;
	}

	.inline-action:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}

	.ledger {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-block-size: 0;
		inline-size: 100%;
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-lg);
		background: var(--color-bg-elevated);
		overflow: hidden;
	}

	.rows {
		/* The rows are a table: the name takes what the columns leave. */
		--ledger-name-basis: 0px;
		flex: 1;
		min-block-size: 0;
		overflow-y: auto;
	}

	.rows:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: calc(var(--focus-ring-offset) * -1);
	}

	.toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 4px 12px;
		min-block-size: 48px;
		padding-block: 8px;
		padding-inline: var(--density-cell-padding-x);
	}

	/* The span fills the left and the scrubber sits right. Compact stacks
	   them, span above scrubber. */
	.pager {
		display: flex;
		align-items: center;
		justify-content: space-between;
		block-size: 56px;
		background: var(--color-bg-surface);
	}

	.pager.compact {
		flex-direction: column;
		align-items: stretch;
		justify-content: center;
		block-size: 92px;
	}

	.totals {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: 2px var(--space-4);
	}

	.figures {
		display: flex;
		gap: var(--space-3);
		margin: 0;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
		white-space: nowrap;
	}

	.why {
		display: grid;
		place-items: center;
		inline-size: 24px;
		block-size: 24px;
		padding: 0;
		border: 0;
		border-radius: var(--radius-sm);
		background: none;
		color: var(--color-text-secondary);
		cursor: pointer;
	}

	.why:hover {
		background: var(--color-bg-hover);
		color: var(--color-text-primary);
	}

	.why:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.note {
		margin: 0;
		padding: 0 var(--density-cell-padding-x) var(--space-3);
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
		text-align: end;
	}

	.amount {
		font-family: var(--type-family-mono);
		font-weight: 500;
	}

	.income {
		color: var(--color-amount-income);
	}

	.expense {
		color: var(--color-amount-expense);
	}

	.count {
		flex: 1 1 auto;
		margin: 0;
		color: var(--color-text-primary);
		font-size: 13px;
		font-weight: 500;
	}

	.count:focus-visible {
		outline: var(--focus-ring-width) solid var(--color-focus-ring);
		outline-offset: var(--focus-ring-offset);
	}

	.columns {
		display: flex;
		align-items: center;
		gap: 12px;
		block-size: 38px;
		padding-inline: var(--density-cell-padding-x);
		background: var(--color-bg-surface);
		color: var(--color-text-secondary);
		font-size: 11px;
		font-weight: 600;
	}

	.col-gap {
		inline-size: 28px;
	}

	.col-description {
		flex: 1 1 auto;
	}

	.col-account {
		inline-size: 220px;
	}

	.col-status {
		inline-size: 92px;
	}

	.col-amount {
		inline-size: 112px;
		text-align: end;
	}

	/* At Medium the columns leave the name about 180, which cuts a bank's line
	   short. The account gives up the room it was not using. */
	@media (min-width: 768px) and (max-width: 1199px) {
		.rows {
			--ledger-account-basis: 168px;
		}

		.col-account {
			inline-size: 168px;
		}
	}

	/* A day repeated at the top of a page is the same heading, not a new one. */
	.day {
		display: flex;
		align-items: center;
		margin: 0;
		block-size: 20px;
		padding-inline: var(--density-cell-padding-x);
		background: var(--color-bg-subtle);
		color: var(--color-text-secondary);
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.02em;
	}

	.sr-only {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		margin: -1px;
		padding: 0;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
</style>
