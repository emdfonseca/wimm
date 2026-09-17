<script lang="ts" module>
	/**
	 * J03.A / 04 · Choose accounts, and J06.A / 02, which is this same screen
	 * reached from a restore.
	 *
	 * **Not a consent step.** By the time it appears the bank has already
	 * granted access, and some banks let a member narrow accounts in their own
	 * screen while others hand over everything. So this asks one question —
	 * who in the household sees each account — and never implies it restricts
	 * the bank.
	 *
	 * Finish is always enabled. Finishing having granted nobody anything is a
	 * legitimate outcome: the accounts are the connecting member's and they can
	 * see them; the household simply sees nothing of this bank yet.
	 */
	export type Level = 'hidden' | 'balance' | 'details';

	export interface ChoiceAccount {
		id: string;
		name: string;
		meta: string;
		balance?: string;
		owned: boolean;
		newlyOffered?: boolean;
	}

	export interface HouseholdMember {
		id: string;
		name: string;
	}
</script>

<script lang="ts">
	import AccountChoiceRow from '../molecules/AccountChoiceRow.svelte';
	import Button from '../atoms/Button.svelte';
	import InfoNotice from '../molecules/InfoNotice.svelte';

	interface Props {
		bankName: string;
		accounts: ChoiceAccount[];
		/** Everyone else in the household. Empty in a household of one. */
		members?: HouseholdMember[];
		/** account id -> member id -> level. Absent pairs are hidden. */
		levels?: Record<string, Record<string, Level>>;
		/** The member owns none of these, so they may look and change nothing. */
		readOnly?: boolean;
		/** Reached by restoring rather than by connecting for the first time. */
		restoring?: boolean;
		onownedchange?: (accountId: string, owned: boolean) => void;
		onlevelchange?: (accountId: string, memberId: string, level: Level) => void;
		onfinish?: () => void;
	}

	let {
		bankName,
		accounts,
		members = [],
		levels = {},
		readOnly = false,
		restoring = false,
		onownedchange,
		onlevelchange,
		onfinish
	}: Props = $props();

	/**
	 * Announced when a level changes, naming the account and where it moved to.
	 * Not a running count: with three levels per member there is no single
	 * number to count, and "Joint account, balance" is what a member needs to
	 * hear back.
	 */
	let announcement = $state('');

	/**
	 * The heading takes focus on arrival. A member reaches this screen by
	 * coming back from their bank, which is a full navigation: without this a
	 * screen reader lands at the top of a page it has already announced, and
	 * the member has no idea the hand-off worked.
	 *
	 * No focus ring: it is tabindex="-1", so nobody arrives here by tabbing.
	 */
	let headingElement = $state<HTMLHeadingElement | null>(null);

	$effect(() => {
		headingElement?.focus();
	});

	const unread = $derived(accounts.filter((a) => !a.owned && !anyGrant(a.id)));
	const newlyOffered = $derived(accounts.some((a) => a.newlyOffered));

	function anyGrant(accountId: string) {
		return Object.values(levels[accountId] ?? {}).some((level) => level !== 'hidden');
	}

	function levelChanged(account: ChoiceAccount, memberId: string, level: Level) {
		const member = members.find((m) => m.id === memberId);
		const said = level === 'hidden' ? 'nothing' : level;
		announcement = `${member?.name ?? 'They'} now sees ${said} of ${account.name}`;
		onlevelchange?.(account.id, memberId, level);
	}

	function ownedChanged(account: ChoiceAccount, owned: boolean) {
		announcement = owned ? `${account.name} is yours` : `${account.name} is nobody's`;
		onownedchange?.(account.id, owned);
	}
</script>

<main class="screen">
	<h1 bind:this={headingElement} tabindex="-1">Who sees these accounts?</h1>

	<p class="lede">
		{#if newlyOffered}
			Your bank offered an account it had not before. It is yours until you say otherwise, and
			nobody else sees it yet.
		{:else if restoring}
			{bankName} is connected again.
		{:else}
			These accounts are yours. Choose who else sees each one, and how much of it — nobody sees
			anything until you say so.
		{/if}
	</p>

	{#if readOnly}
		<InfoNotice title="These are not yours to change">
			You can see what you have been given of {bankName}'s accounts. Whoever owns them decides who
			else sees them.
		</InfoNotice>
	{/if}

	<section class="accounts" aria-label="{bankName} accounts">
		<h2>{bankName}</h2>

		{#each accounts as account (account.id)}
			<AccountChoiceRow
				name={account.name}
				meta={account.meta}
				balance={account.balance}
				owned={account.owned}
				newlyOffered={account.newlyOffered}
				disabled={readOnly}
				{members}
				levels={levels[account.id] ?? {}}
				onownedchange={(owned) => ownedChanged(account, owned)}
				onlevelchange={(memberId, level) => levelChanged(account, memberId, level)}
			/>
		{/each}
	</section>

	<!-- Names the account and the level it moved to, because that is what a
	     member needs to hear back. -->
	<p class="sr-only" role="status" aria-live="polite">{announcement}</p>

	<div class="actions">
		<Button onclick={onfinish}>Finish</Button>
		<p class="helper">
			{#if restoring}
				<!-- On a restore the reassurance belongs here rather than in the lede,
				     which the newly-offered account needs (frame J06.A / 02). -->
				The accounts you already chose kept their owners and levels.
			{:else if unread.length > 0}
				{unread.length}
				{unread.length === 1 ? 'account is' : 'accounts are'} nobody's. wimm will not read a
				balance for {unread.length === 1 ? 'it' : 'them'}.
			{:else}
				Leaving everything as it is keeps these accounts to yourself. You can change any of this
				later.
			{/if}
		</p>
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
	}

	h1 {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-display);
		font-size: var(--type-size-page-title);
		font-weight: 600;
	}
	/* No ring: the heading takes focus to place a screen reader, not to show a
	   keyboard user where they are. A scoped focus-visible rule would outrank
	   base.css's attribute selector and quietly win. */
	h1:focus,
	h1:focus-visible {
		outline: none;
	}

	h2 {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		font-weight: 600;
	}

	.lede,
	.helper {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.helper {
		font-size: var(--type-size-body-sm);
	}

	.accounts {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.actions {
		display: flex;
		flex-direction: column;
		gap: 6px;
		align-items: flex-start;
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
