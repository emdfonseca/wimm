<script lang="ts" module>
	/**
	 * J03.A / 04 · Choose accounts, and J06.A / 02, which is this same screen
	 * reached from a restore. Gains the owner set, the last-owner refusal, and
	 * leaving an account out / bringing it back — `J09.A / 01` through `06`.
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
	 *
	 * Ownership can never fall to zero (ADR 0022). Ticking a co-owner reports
	 * only that member, never a replacement list — the defect the single
	 * "Mine" checkbox had — and a change that would leave an account with
	 * nobody is refused inline, beside the row, naming leaving the account out
	 * as the way to the same end.
	 */
	export type Level = 'hidden' | 'balance' | 'details';

	export interface ChoiceAccount {
		id: string;
		/** The household's name where one is set, else the bank's own. */
		name: string;
		/** The raw household name, empty when nobody has set one — what the
		 *  rename field starts from, as opposed to `name`'s already-folded
		 *  display value. */
		householdName?: string;
		/** The bank's own name, always shown beneath the field when renaming
		 *  (ADR 0022) — kept even once a household name replaces it above. */
		bankAccountName: string;
		bankName: string;
		meta: string;
		balance?: string;
		owned: boolean;
		newlyOffered?: boolean;
		/** The account is left out of wimm; the row offers to bring it back. */
		leftOut?: boolean;
	}

	export interface HouseholdMember {
		id: string;
		name: string;
	}
</script>

<script lang="ts">
	import AccountChoiceRow, { type OtherMember } from '../molecules/AccountChoiceRow.svelte';
	import AccountName from '../molecules/AccountName.svelte';
	import Button from '../atoms/Button.svelte';
	import InfoNotice from '../molecules/InfoNotice.svelte';
	import Notice from '../atoms/Notice.svelte';
	import LeftOutDialog, { type Grantee } from '../molecules/LeftOutDialog.svelte';
	import Page from '../templates/Page.svelte';

	interface Props {
		bankName: string;
		accounts: ChoiceAccount[];
		/** Everyone else in the household. Empty in a household of one. */
		members?: HouseholdMember[];
		/** account id -> member id -> whether that member already owns it. */
		owners?: Record<string, Record<string, boolean>>;
		/** account id -> member id -> level. Meaningful only for a member who
		 *  is not an owner of that account; absent pairs are hidden. */
		levels?: Record<string, Record<string, Level>>;
		/** The member owns none of these, so they may look and change nothing. */
		readOnly?: boolean;
		/** Reached by restoring rather than by connecting for the first time. */
		restoring?: boolean;
		/** Set after the database refused to leave an account with no owner —
		 *  the refusal appears inline on that one row. */
		refusedAccountId?: string | null;
		onownedchange?: (accountId: string, owned: boolean) => void;
		/** One member's owner checkbox changed. Add or remove just that
		 *  member — never a replacement list. */
		onownerchange?: (accountId: string, memberId: string, isOwner: boolean) => void;
		onlevelchange?: (accountId: string, memberId: string, level: Level) => void;
		onleaveout?: (accountId: string) => void;
		onbringback?: (accountId: string) => void;
		onrename?: (accountId: string, householdName: string) => void;
		onfinish?: () => void;
	}

	let {
		bankName,
		accounts,
		members = [],
		owners = {},
		levels = {},
		readOnly = false,
		restoring = false,
		refusedAccountId = null,
		onownedchange,
		onownerchange,
		onlevelchange,
		onleaveout,
		onbringback,
		onrename,
		onfinish
	}: Props = $props();

	/** Which account's name is being edited, ephemeral and on the row it
	 *  belongs to — never more than one at a time. */
	let renaming = $state<string | null>(null);

	/**
	 * Announced when a level or an owner changes, naming the account and what
	 * moved. Not a running count: there is no single number that describes
	 * three levels and an owner set, and "Joint account, balance" is what a
	 * member needs to hear back.
	 */
	let announcement = $state('');

	const unread = $derived(accounts.filter((a) => !a.owned && !a.leftOut && !anyGrant(a.id)));
	const newlyOffered = $derived(accounts.some((a) => a.newlyOffered));

	function anyGrant(accountId: string) {
		return Object.values(levels[accountId] ?? {}).some((level) => level !== 'hidden');
	}

	function othersFor(account: ChoiceAccount): OtherMember[] {
		const accountOwners = owners[account.id] ?? {};
		const accountLevels = levels[account.id] ?? {};
		return members.map((m) => ({
			id: m.id,
			name: m.name,
			isOwner: accountOwners[m.id] ?? false,
			level: accountLevels[m.id] ?? 'hidden'
		}));
	}

	function grantees(account: ChoiceAccount): Grantee[] {
		const accountOwners = owners[account.id] ?? {};
		const accountLevels = levels[account.id] ?? {};
		return members
			.filter((m) => !accountOwners[m.id])
			.map((m) => ({ name: m.name, level: accountLevels[m.id] }))
			.filter((g): g is Grantee => g.level === 'balance' || g.level === 'details');
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

	function ownerChanged(account: ChoiceAccount, memberId: string, isOwner: boolean) {
		const member = members.find((m) => m.id === memberId);
		announcement = isOwner
			? `${member?.name ?? 'They'} now owns ${account.name}`
			: `${member?.name ?? 'They'} no longer owns ${account.name}`;
		onownerchange?.(account.id, memberId, isOwner);
	}

	/** The dialog currently open, or none. Only one can be open at a time. */
	let dialog = $state<{ account: ChoiceAccount; mode: 'leave-out' | 'bring-back' } | null>(null);

	function openDialog(account: ChoiceAccount) {
		dialog = { account, mode: account.leftOut ? 'bring-back' : 'leave-out' };
	}

	function confirmDialog() {
		if (!dialog) return;
		const { account, mode } = dialog;
		if (mode === 'leave-out') onleaveout?.(account.id);
		else onbringback?.(account.id);
	}
</script>

<Page title="Who sees these accounts?" focusHeading>
	<p class="lede">
		{#if newlyOffered}
			Your bank offered an account it had not before. It is yours until you say otherwise, and
			nobody else sees it yet.
		{:else if restoring}
			{bankName} is connected again.
		{:else}
			These accounts are yours. Choose who else sees each one, and how much of it. Nobody sees
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
			<div class="account-block">
				<AccountChoiceRow
					name={account.name}
					meta={account.meta}
					balance={account.balance}
					owned={account.owned}
					newlyOffered={account.newlyOffered}
					leftOut={account.leftOut}
					disabled={readOnly}
					others={othersFor(account)}
					onownedchange={(owned) => ownedChanged(account, owned)}
					onownerchange={(memberId, isOwner) => ownerChanged(account, memberId, isOwner)}
					onlevelchange={(memberId, level) => levelChanged(account, memberId, level)}
					onleaveout={() => openDialog(account)}
				/>

				{#if renaming === account.id}
					<AccountName
						householdName={account.householdName ?? ''}
						bankName={account.bankName}
						bankAccountName={account.bankAccountName}
						onSave={(householdName) => {
							onrename?.(account.id, householdName);
							renaming = null;
						}}
					/>
				{:else}
					<button type="button" class="rename-trigger" onclick={() => (renaming = account.id)}>
						Rename
					</button>
				{/if}

				{#if refusedAccountId === account.id}
					<div class="refusal">
						<Notice title="An account has to belong to somebody">
							You are its only owner, so this would leave it with nobody. Leave it out of wimm
							instead: nothing already read is deleted, and you can bring it back whenever you like.
						</Notice>
						<Button variant="secondary" size="sm" onclick={() => openDialog(account)}>
							Leave it out
						</Button>
					</div>
				{/if}
			</div>
		{/each}
	</section>

	{#if dialog}
		<LeftOutDialog
			open={true}
			mode={dialog.mode}
			accountName={dialog.account.name}
			grantees={grantees(dialog.account)}
			onconfirm={confirmDialog}
			oncancel={() => (dialog = null)}
		/>
	{/if}

	<!-- Names the account and what moved, because that is what a member needs
	     to hear back. -->
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
				{unread.length === 1 ? 'account is' : 'accounts are'} nobody's. wimm will not read a balance for
				{unread.length === 1 ? 'it' : 'them'}.
			{:else}
				Leaving everything as it is keeps these accounts to yourself. You can change any of this
				later.
			{/if}
		</p>
	</div>
</Page>

<style>
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

	.account-block {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.rename-trigger {
		align-self: flex-start;
		padding: 0;
		border: none;
		background: none;
		color: var(--color-accent);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 500;
		cursor: pointer;
	}

	.rename-trigger:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.refusal {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 8px;
		padding-inline-start: 12px;
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
