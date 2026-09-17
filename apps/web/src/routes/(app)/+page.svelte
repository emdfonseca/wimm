<script lang="ts">
	import { untrack } from 'svelte';
	import { AccountsOverview, DisconnectBankDialog } from '@wimm/ui';
	import { afterNavigate, goto, invalidate } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { data } = $props();

	let refreshing = $state(false);

	/**
	 * What the last hand-off produced. Held here because the load spends it on
	 * read: re-reading this page when the arrival's balances land would clear
	 * the notice a second or two after it appeared. It is assigned on
	 * navigation, which is the only time a new one can exist.
	 */
	// untrack because capturing the initial value is the point: re-deriving it
	// would put the notice back under the load that spends it.
	let outcome = $state(untrack(() => data.outcome));
	let outcomeBank = $state(untrack(() => data.outcomeBank));
	let outcomeAccessEndsOn = $state(untrack(() => data.outcomeAccessEndsOn));
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

	/**
	 * Read the balances behind the arrival, and again when the member asks.
	 *
	 * `read-balances` is the arrival; `refresh` is the member asking in as many
	 * words. Neither runs in the load function, because a load holds the
	 * navigation open until it returns.
	 */
	async function read(action: 'read-balances' | 'refresh') {
		if (refreshing) return;
		refreshing = true;
		try {
			await act({ action });
			// Whether or not every bank answered: wimmd keeps the readings it
			// could not renew, and the reload shows what it has.
			await invalidate('wimm:accounts');
		} finally {
			refreshing = false;
		}
	}

	/**
	 * Once per arrival, and not once per render.
	 *
	 * `afterNavigate` fires on the initial load and on every navigation after
	 * it, and not on an invalidation — which is what stops the read that
	 * re-reads this page from triggering itself.
	 */
	afterNavigate(() => {
		outcome = data.outcome;
		outcomeBank = data.outcomeBank;
		outcomeAccessEndsOn = data.outcomeAccessEndsOn;
		void read('read-balances');
	});

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
		await invalidate('wimm:accounts');
	}

</script>

<AccountsOverview
	outcome={disconnected ? 'disconnected' : outcome}
	outcomeBank={disconnected?.bankName ?? outcomeBank}
	outcomeAccountCount={disconnected?.accountCount ?? 0}
	outcomeAccessEndsOn={outcomeAccessEndsOn ?? ''}
	banks={data.banks}
	accounts={data.accounts}
	totals={data.totals}
	problems={data.problems}
	{refreshing}
	onrefresh={() => read('refresh')}
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
