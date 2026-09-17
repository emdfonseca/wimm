<script lang="ts">
	import { AccountsOverview, DisconnectBankDialog } from '@wimm/ui';
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { data } = $props();

	let refreshing = $state(false);
	let disconnecting: { connectionId: string; bankName: string; accountCount: number } | null =
		$state(null);

	/**
	 * A bank that has just been disconnected — frame J05.A / 02.
	 *
	 * Held here rather than carried in the URL like the other outcomes:
	 * disconnecting never leaves the page, so there is no return to put an
	 * outcome on.
	 */
	let disconnected: { bankName: string; accountCount: number } | null = $state(null);

	async function act(body: Record<string, string>) {
		const response = await fetch('/api/banking', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(body)
		});
		return (await response.json()) as { ok: boolean; handoffUrl?: string; reason?: string };
	}

	async function refresh() {
		refreshing = true;
		try {
			await act({ action: 'refresh' });
			// Whether or not every bank answered: wimmd keeps the readings it
			// could not renew, and the reload shows what it has.
			await invalidateAll();
		} finally {
			refreshing = false;
		}
	}

	async function restore(connectionId: string) {
		// The name travels with the request so the return can name the bank:
		// what comes back from a bank carries only the state value.
		const bankName = data.banks.find((bank) => bank.connectionId === connectionId)?.bankName ?? '';
		const result = await act({ action: 'restore', connectionId, bankName });
		// A full navigation, not goto: the destination is the bank.
		if (result.handoffUrl) location.href = result.handoffUrl;
	}

	async function confirmDisconnect() {
		const target = disconnecting;
		disconnecting = null;
		if (!target) return;

		await act({ action: 'disconnect', connectionId: target.connectionId });
		disconnected = { bankName: target.bankName, accountCount: target.accountCount };
		await invalidateAll();
	}

</script>

<AccountsOverview
	outcome={disconnected ? 'disconnected' : data.outcome}
	outcomeBank={disconnected?.bankName ?? data.outcomeBank}
	outcomeAccountCount={disconnected?.accountCount ?? 0}
	outcomeAccessEndsOn={data.outcomeAccessEndsOn ?? ''}
	banks={data.banks}
	accounts={data.accounts}
	totals={data.totals}
	problems={data.problems}
	{refreshing}
	onrefresh={refresh}
	onconnect={() => goto(resolve('/(app)/connect'))}
	onrestore={restore}
	onmanage={(bank) =>
		goto(
			resolve('/(app)/connect/[bank]/accounts/[connection]', {
				bank: bank.bankId,
				connection: bank.connectionId
			})
		)}
	ondisconnect={(bank) =>
		(disconnecting = {
			connectionId: bank.connectionId,
			bankName: bank.bankName,
			accountCount: bank.accountCount
		})}
/>

{#if disconnecting}
	<DisconnectBankDialog
		open={true}
		bankName={disconnecting.bankName}
		accountCount={disconnecting.accountCount}
		onconfirm={confirmDisconnect}
		oncancel={() => (disconnecting = null)}
	/>
{/if}
