<script lang="ts">
	import { Overview } from '@wimm/ui';
	import { afterNavigate, goto, invalidate } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { data } = $props();

	/**
	 * Balances are read when a member arrives (`banking/household-accounts`),
	 * and that arrival is Overview: the screen a member lands on after signing
	 * in. Accounts no longer reads on its own load, only on its own Refresh
	 * button — reading here too would read every account a second time on
	 * every visit there.
	 *
	 * Not inside the load: a load blocks navigation until it returns, and
	 * reading balances is a round trip per account, so reading here means
	 * landing on Overview and watching it until every bank has answered. The
	 * figures already held render with the time they were read, which is the
	 * contract the spec actually makes.
	 */
	async function readBalances() {
		await fetch('/api/banking', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ action: 'read-balances' })
		});
		// Whether or not every bank answered: wimmd keeps the readings it could
		// not renew, and the reload shows what it has.
		await invalidate('wimm:accounts');
	}

	/**
	 * Once per arrival, and not once per render. `afterNavigate` fires on the
	 * initial load and on every navigation after it, and not on an
	 * invalidation — which is what stops this read from triggering itself.
	 */
	afterNavigate(() => {
		void readBalances();
	});
</script>

<Overview
	hasAccounts={data.hasAccounts}
	ownsNothing={data.ownsNothing}
	currencies={data.currencies}
	accounts={data.accounts}
	recentTransactions={data.recentTransactions}
	ongotoaccounts={() => goto(resolve('/(app)/accounts'))}
	onseeall={() => goto(resolve('/(app)/transactions'))}
/>
