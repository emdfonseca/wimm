<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { invalidateAll } from '$app/navigation';
	import { ChooseAccountsScreen, type Level } from '@wimm/ui';

	let { data } = $props();

	/** What the last change did, when it did not work. */
	let problem = $state('');
	/** Set when the database refused to leave an account with no owner —
	 *  the screen names it inline, beside that one row (ADR 0022). */
	let refusedAccountId = $state<string | null>(null);

	type Owners = Record<string, Record<string, boolean>>;

	/**
	 * Each change is posted on its own rather than batched behind Finish.
	 *
	 * A member who sets three levels and closes the tab has made three
	 * decisions, and the alternative loses all of them — which matters more
	 * here than elsewhere because the bank has already granted access and
	 * coming back costs another round trip through it.
	 */
	async function post(
		action: 'owners' | 'level' | 'leaveOut' | 'bringBack' | 'rename',
		body: Record<string, string | string[]>,
		accountId: string
	) {
		const form = new FormData();
		for (const [key, value] of Object.entries(body)) {
			if (Array.isArray(value)) {
				for (const item of value) form.append(key, item);
			} else {
				form.set(key, value);
			}
		}

		const response = await fetch(`?/${action}`, { method: 'POST', body: form });

		// The response was thrown away here, so a refused change looked exactly
		// like an applied one and the member had no way to know. It still is not
		// deserialised — invalidateAll re-reads the truth from the server — but
		// a failure now says so, and the last-owner refusal names the account.
		refusedAccountId = null;
		if (!response.ok) {
			if (response.status === 409) {
				refusedAccountId = accountId;
				problem = '';
			} else {
				problem =
					response.status === 403
						? 'That account is not yours to change.'
						: 'That change could not be saved. Nothing was altered.';
			}
		} else {
			problem = '';
		}

		await invalidateAll();
	}

	/** The account's current owners, plus or minus the connecting member — who
	 *  is not in the household member list `data.owners` keys off, so their id
	 *  is threaded in separately. */
	function ownerIds(owners: Owners, accountId: string, memberOwned: boolean): string[] {
		const ids = Object.entries(owners[accountId] ?? {})
			.filter(([, isOwner]) => isOwner)
			.map(([id]) => id);
		return memberOwned ? [...ids, data.currentMemberId] : ids;
	}
</script>

{#if problem}
	<p class="problem" role="alert">{problem}</p>
{/if}

<ChooseAccountsScreen
	bankName={data.bankName}
	accounts={data.accounts}
	members={data.members}
	owners={data.owners as Owners}
	levels={data.levels as Record<string, Record<string, Level>>}
	readOnly={data.readOnly}
	{refusedAccountId}
	onownedchange={(accountId, owned) => {
		const others = Object.entries((data.owners as Owners)[accountId] ?? {})
			.filter(([, isOwner]) => isOwner)
			.map(([id]) => id);
		post('owners', { accountId, memberIds: owned ? [...others, data.currentMemberId] : others }, accountId);
	}}
	onownerchange={(accountId, memberId, isOwner) => {
		const owners = data.owners as Owners;
		const currentlyOwned = Boolean(
			data.accounts.find((a: { id: string }) => a.id === accountId)?.owned
		);
		const without = ownerIds(owners, accountId, currentlyOwned).filter((id) => id !== memberId);
		const memberIds = isOwner ? [...without, memberId] : without;
		post('owners', { accountId, memberIds }, accountId);
	}}
	onlevelchange={(accountId, memberId, level) =>
		post('level', { accountId, memberId, level }, accountId)}
	onleaveout={(accountId) => post('leaveOut', { accountId }, accountId)}
	onbringback={(accountId) => post('bringBack', { accountId }, accountId)}
	onrename={(accountId, householdName) => post('rename', { accountId, householdName }, accountId)}
	onfinish={() => goto(resolve('/(app)/accounts'))}
/>

<style>
	.problem {
		margin: 0;
		padding: 12px 16px;
		background: var(--color-feedback-danger-bg);
		color: var(--color-feedback-danger);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}
</style>
