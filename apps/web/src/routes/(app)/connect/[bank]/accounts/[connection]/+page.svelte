<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { invalidateAll } from '$app/navigation';
	import { ChooseAccountsScreen, type Level } from '@wimm/ui';

	let { data } = $props();

	/** What the last change did, when it did not work. */
	let problem = $state('');

	/**
	 * Each change is posted on its own rather than batched behind Finish.
	 *
	 * A member who sets three levels and closes the tab has made three
	 * decisions, and the alternative loses all of them — which matters more
	 * here than elsewhere because the bank has already granted access and
	 * coming back costs another round trip through it.
	 */
	async function post(action: 'owners' | 'level', body: Record<string, string | string[]>) {
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
		// a failure now says so.
		if (!response.ok) {
			problem =
				response.status === 403
					? 'That account is not yours to change.'
					: 'That change could not be saved. Nothing was altered.';
		} else {
			problem = '';
		}

		await invalidateAll();
	}
</script>

{#if problem}
	<p class="problem" role="alert">{problem}</p>
{/if}

<ChooseAccountsScreen
	bankName={data.bankName}
	accounts={data.accounts}
	members={data.members}
	levels={data.levels as Record<string, Record<string, Level>>}
	readOnly={data.readOnly}
	onownedchange={(accountId, owned) =>
		post('owners', { accountId, memberIds: owned ? [data.currentMemberId] : [] })}
	onlevelchange={(accountId, memberId, level) =>
		post('level', { accountId, memberId, level })}
	onfinish={() => goto(resolve('/(app)'))}
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
