<script lang="ts" module>
	/**
	 * J04 · Overview. The screen a member lands on after signing in, and the
	 * only destination in this change.
	 *
	 * Everything here is **that member's own view**: the accounts they may see,
	 * and a total of exactly those. Two members of one household land on this
	 * screen and correctly see different numbers, and neither is told what the
	 * other sees.
	 */
	export interface OverviewAccount {
		id: string;
		/** Which connection this account belongs to, so it sits under its bank. */
		connectionId?: string;
		name: string;
		bank: string;
		/** Present only where the member may see it. */
		numberSuffix?: string;
		balance?: string;
		readAt?: string;
		stale?: boolean;
		notUpdating?: boolean;
		negative?: boolean;
	}

	export interface CurrencyTotal {
		/** Already formatted, carrying its currency. */
		total: string;
		currency: string;
	}

	/**
	 * What the last hand-off produced, for a member who has just come back
	 * from one. Distinct from BankProblem: that is a bank wimm already has and
	 * could not read, this is an attempt to add one that did not land.
	 */
	export type Outcome =
		| 'connected'
		/** Access renewed at a bank already connected — frame J06.A / 03. */
		| 'restored'
		| 'disconnected'
		| 'declined'
		| 'no-accounts'
		| 'bank-unavailable'
		| 'already-connected';

	/**
	 * A bank the household has connected, as the viewing member sees it.
	 *
	 * Overview showed accounts and nothing about where they came from, which
	 * left two specified things missing and one capability unreachable: who
	 * connected a bank, when its access ends, and any way to open the chooser
	 * or disconnect it.
	 */
	export interface ConnectedBank {
		connectionId: string;
		bankId: string;
		bankName: string;
		/** Whose consent holds it open, and who will have to restore it. */
		connectedBy: string;
		/** Already formatted, e.g. "16 December 2026". */
		accessEndsOn: string;
		/** False once access has run out. */
		live: boolean;
		accountCount: number;
		/** The member may change owners and levels on this bank. */
		manageable: boolean;
	}

	export interface BankProblem {
		connectionId: string;
		bankName: string;
		kind: 'unreachable' | 'rate-limited' | 'access-ended';
		/** Only for rate-limited, and only when the bank said. */
		retryAfter?: string;
	}
</script>

<script lang="ts">
	import AccountRow from '../molecules/AccountRow.svelte';
	import Button from '../atoms/Button.svelte';
	import EmptyState from '../molecules/EmptyState.svelte';
	import ErrorNotice from '../molecules/ErrorNotice.svelte';
	import InfoNotice from '../molecules/InfoNotice.svelte';

	interface Props {
		outcome?: Outcome;
		/** The bank the outcome is about. The frames name it in every one. */
		outcomeBank?: string;
		/** How many accounts went, which the disconnected notice counts. */
		outcomeAccountCount?: number;
		/** Already formatted. The restored notice names how long access runs. */
		outcomeAccessEndsOn?: string;
		banks?: ConnectedBank[];
		accounts?: OverviewAccount[];
		totals?: CurrencyTotal[];
		problems?: BankProblem[];
		refreshing?: boolean;
		onrefresh?: () => void;
		onconnect?: () => void;
		onrestore?: (connectionId: string) => void;
		/** Open the chooser for this bank. */
		onmanage?: (bank: ConnectedBank) => void;
		/** Ask to disconnect it. Confirmation is the caller's. */
		ondisconnect?: (bank: ConnectedBank) => void;
	}

	let {
		outcome,
		outcomeBank = 'That bank',
		outcomeAccountCount = 0,
		outcomeAccessEndsOn = '',
		banks = [],
		accounts = [],
		totals = [],
		problems = [],
		refreshing = false,
		onrefresh,
		onconnect,
		onrestore,
		onmanage,
		ondisconnect
	}: Props = $props();

	/** Every account id that a bank card above will render. */
	const grouped = $derived(
		new Set(
			banks.flatMap((bank) =>
				accounts.filter((a) => a.connectionId === bank.connectionId).map((a) => a.id)
			)
		)
	);

	/**
	 * The bank card behind a problem. The notice names who granted the access
	 * and when it ran out, and both are on the card rather than on the
	 * failure — a failure says which connection, not who is holding it open.
	 */
	const bankOf = (problem: BankProblem) =>
		banks.find((bank) => bank.connectionId === problem.connectionId);

	const ended = $derived(problems.filter((p) => p.kind === 'access-ended'));
	const transient = $derived(problems.filter((p) => p.kind !== 'access-ended'));

	/**
	 * A member returning to a failure lands on the banner rather than at the
	 * top of a page whose figures look unchanged — which is exactly what a
	 * stale reading looks like.
	 */
	let banner = $state<HTMLElement | null>(null);

	$effect(() => {
		if (problems.length > 0) banner?.focus();
	});

	/**
	 * What refreshing produced, announced rather than only drawn: the figures
	 * change in place, and a member not looking at them has no other way to
	 * know the request finished.
	 */
	const refreshOutcome = $derived.by(() => {
		if (refreshing) return 'Refreshing balances';
		if (problems.length === 0) return '';

		// The same words the notice uses, because a member hearing one and
		// reading the other must not be told two different things.
		const limited = problems.find((p) => p.kind === 'rate-limited');
		if (limited) {
			return limited.retryAfter
				? `Balances cannot be refreshed yet. ${limited.bankName} limits how often wimm may read balances. The next refresh is possible after ${limited.retryAfter}.`
				: `Balances cannot be refreshed yet. ${limited.bankName} limits how often wimm may read balances.`;
		}

		const names = problems.map((p) => p.bankName).join(', ');
		return `${names} could not be updated. The balances shown are the last ones wimm read.`;
	});
</script>

<!-- Copy from frames J03.B / J03.C / J03.D / J05.A / 02 / J06.A / 03, with
     the bank where the frame draws Monzo. Every one of these already had
     designed wording, and I wrote my own beside it. -->
{#snippet outcomeNotice()}
	{#if outcome === 'connected'}
		<InfoNotice title="{outcomeBank} is connected">
			Its accounts are below. They are yours, and nobody else in the household sees them yet.
		</InfoNotice>
	{:else if outcome === 'declined'}
		<InfoNotice title="{outcomeBank} was not connected">
			Access was not granted at {outcomeBank}, so nothing was added to the household. You can try
			{outcomeBank} again, or pick a different bank.
		</InfoNotice>
	{:else if outcome === 'no-accounts'}
		<ErrorNotice title="No accounts came back from {outcomeBank}" live="polite">
			{outcomeBank} granted access but offered no accounts wimm can read. The bank was not
			connected. If you hold accounts at {outcomeBank}, check whether they are covered by the
			access you granted.
		</ErrorNotice>
	{:else if outcome === 'already-connected'}
		<InfoNotice title="{outcomeBank} is already connected">
			The link you came back on had been used. Its accounts are below.
		</InfoNotice>
	{:else if outcome === 'disconnected'}
		<InfoNotice title="{outcomeBank} was disconnected">
			Its {outcomeAccountCount} accounts are no longer shown to anyone in the household. If {outcomeBank}
			still lists wimm among the apps it shares data with, remove it there as well.
		</InfoNotice>
	{:else if outcome === 'restored'}
		<InfoNotice title="{outcomeBank} is updating again">
			Access runs until {outcomeAccessEndsOn}. wimm will say so here again when it is close to
			running out.
		</InfoNotice>
	{:else if outcome}
		<ErrorNotice title="{outcomeBank} could not be reached" live="polite">
			wimm could not finish connecting, so nothing was added. This is not something you did. Try
			again in a few minutes.
		</ErrorNotice>
	{/if}
{/snippet}

<main class="screen">
	<header class="head">
		<h1 tabindex="-1">Overview</h1>
		{#if accounts.length > 0}
			<!-- Primary in the frame: it is the one thing a member comes back to
			     this screen to do. -->
			<Button onclick={onrefresh} disabled={refreshing}>
				{refreshing ? 'Refreshing…' : 'Refresh balances'}
			</Button>
		{/if}
	</header>

	<!-- Under the page header, not above it: J04.B / 01 draws Overview, then
	     Refresh balances, then the notice, then the total. -->
	{@render outcomeNotice()}

	{#if accounts.length === 0}
		<!-- No total over nothing: a total of zero would claim a household has
		     no money rather than that wimm has not been told. -->
		<EmptyState title="No banks connected">
			Connect a bank and wimm shows the household's accounts and balances here.
		</EmptyState>
		<Button onclick={onconnect}>Connect a bank</Button>
	{:else}
		<!-- One focus target for every banner, because a member returning to a
		     failure needs to land on the explanation and not on the figures,
		     which look unchanged when they are stale. -->
		{#if problems.length > 0}
			<div class="banner" tabindex="-1" bind:this={banner}>
				<!-- Consent having run out is not a failure to retry: the member
				     restores it, which is the whole flow again. -->
				{#each ended as problem (problem.connectionId)}
					<InfoNotice title="{problem.bankName} has stopped updating">
						The access {bankOf(problem)?.connectedBy ?? 'someone else'} granted ran out on {bankOf(
							problem
						)?.accessEndsOn ?? 'an earlier date'}, so these balances are the last ones read.
						Restoring takes a moment at {problem.bankName} and keeps the accounts the household is
						already sharing.
					</InfoNotice>
					<Button variant="secondary" onclick={() => onrestore?.(problem.connectionId)}>
						Restore access
					</Button>
				{/each}

				{#each transient as problem (problem.connectionId)}
					{#if problem.kind === 'rate-limited'}
						<!-- The bank's own cap, not a wimm one, so the notice says whose
						     limit it is and when it lifts rather than "try later". -->
						<ErrorNotice title="Balances cannot be refreshed yet" live="polite">
							{#if problem.retryAfter}
								{problem.bankName} limits how often wimm may read balances. The next refresh is
								possible after {problem.retryAfter}. The balances below are the ones already read, at
								the times shown.
							{:else}
								{problem.bankName} limits how often wimm may read balances, and did not say when the
								next one is possible. The balances below are the ones already read, at the times
								shown.
							{/if}
						</ErrorNotice>
					{:else}
						<ErrorNotice title="{problem.bankName} could not be updated" live="polite">
							{problem.bankName} did not answer. The balances below are the last ones wimm read.
						</ErrorNotice>
					{/if}
				{/each}
			</div>
		{/if}

		{#if totals.length > 0}
			<!-- Frame J05.A / 01: a labelled tile at 300, then a spacer. Not a
			     bare figure. -->
			<section class="totals" aria-label="Totals">
				<div class="tiles">
					{#each totals as total (total.currency)}
						<div class="total">
							<span class="total-label">
								{totals.length > 1
									? `Household total · ${total.currency}`
									: 'Household total'}
							</span>
							<span class="total-value">{total.total}</span>
						</div>
					{/each}
				</div>
				{#if totals.length > 1}
					<!-- wimm holds no rates, and a converted figure would invent the
					     number a household trusts most. -->
					<p class="no-conversion">Shown per currency. wimm does not convert between them.</p>
				{/if}
			</section>
		{/if}

		<!-- One section per bank: a card header naming it, who connected it and
		     when its access ends, a Disconnect control, then that bank's rows.
		     The frame nests accounts under their bank; a flat list threw that
		     away and left nothing saying where an account came from. -->
		{#each banks as bank (bank.connectionId)}
			<section class="bank" aria-label={bank.bankName}>
				<header class="card-header">
					<span class="bank-name">{bank.bankName}</span>
					<span class="bank-who">
						Connected by {bank.connectedBy} ·
						{bank.live ? `access ends ${bank.accessEndsOn}` : `access ran out ${bank.accessEndsOn}`}
					</span>
					<span class="spacer"></span>
					{#if !bank.live}
						<Button variant="secondary" onclick={() => onrestore?.(bank.connectionId)}>
							Restore access
						</Button>
					{/if}
					{#if bank.manageable}
						<Button variant="ghost" onclick={() => onmanage?.(bank)}>Who sees these</Button>
					{/if}
					<Button variant="secondary" onclick={() => ondisconnect?.(bank)}>Disconnect</Button>
				</header>

				<div class="rows">
					{#each accounts.filter((a) => a.connectionId === bank.connectionId) as account (account.id)}
						<AccountRow
							name={account.name}
							bank={account.bank}
							numberSuffix={account.numberSuffix}
							balance={account.balance}
							readAt={account.readAt}
							stale={account.stale}
							notUpdating={account.notUpdating}
							negative={account.negative}
						/>
					{/each}
				</div>
			</section>
		{/each}

		<!-- Anything not rendered above: an account with no bank behind it, and
		     an account whose bank is missing from `banks`.
		     
		     The second case is the important one. When this screen stopped
		     being passed `banks`, every account had a connectionId matching
		     nothing, so none rendered here either and the page came up empty.
		     A row the member may see is never dropped because its grouping is
		     absent — it falls back to a plain row. -->
		{#each accounts.filter((a) => !a.connectionId || !grouped.has(a.id)) as account (account.id)}
			<AccountRow
				name={account.name}
				bank={account.bank}
				numberSuffix={account.numberSuffix}
				balance={account.balance}
				readAt={account.readAt}
				stale={account.stale}
				notUpdating={account.notUpdating}
				negative={account.negative}
			/>
		{/each}

		<div class="actions">
			<Button variant="secondary" onclick={onconnect}>Connect another bank</Button>
		</div>
		<!-- The outcome of a refresh, for a member who is not watching the
		     figures change. -->
		<p class="sr-only" role="status" aria-live="polite">{refreshOutcome}</p>
	{/if}
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
	}

	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
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

	/* Frame J05.A / 01: a labelled tile at 300, in a row with a spacer. */
	.totals {
		display: flex;
		flex-direction: column;
		gap: 8px;
		align-items: flex-start;
	}

	.tiles {
		display: flex;
		gap: 16px;
		flex-wrap: wrap;
	}

	.total {
		display: flex;
		flex-direction: column;
		gap: 4px;
		inline-size: 300px;
		max-inline-size: 100%;
		padding: 16px;
		background: var(--color-bg-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-md);
	}

	.total-label {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	.total-value {
		color: var(--color-text-primary);
		font-family: var(--type-family-mono);
		font-size: var(--type-size-display-sm);
		font-weight: 600;
	}

	.no-conversion {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	/* A bank and its accounts are one block: the card header names the bank,
	   the rows beneath it are that bank's. */
	/* Frame J05.A / 01: one card per bank, its header and its rows inside the
	   same border. The rows were floating on the page background with no card
	   and no header at all. */
	.bank {
		display: flex;
		flex-direction: column;
		background: var(--color-bg-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-md);
		overflow: hidden;
	}

	/* Height 48 and 16 of horizontal padding, from the frame. */
	.card-header {
		display: flex;
		align-items: center;
		gap: 12px;
		min-block-size: 48px;
		padding-inline: 16px;
	}

	.bank-name {
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		font-weight: 700;
	}

	.bank-who {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	/* The frame's own spacer: it pushes the controls to the end rather than
	   justify-content doing it, so the two texts stay together on the left. */
	.spacer {
		flex: 1 1 auto;
	}

	.rows {
		display: flex;
		flex-direction: column;
	}

	/* A hairline between rows, and none above the first: the card's own border
	   is the line there. */
	.rows > :global(* + *) {
		border-block-start: 1px solid var(--color-border-subtle);
	}

	.card-header {
		border-block-end: 1px solid var(--color-border-subtle);
	}

	@media (max-width: 599px) {
		.card-header {
			flex-wrap: wrap;
			padding-block: 8px;
		}
		.total {
			inline-size: 100%;
		}
	}

	.actions {
		display: flex;
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
