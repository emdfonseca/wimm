<script lang="ts">
	import { TransactionsScreen } from '@wimm/ui';
	import { afterNavigate, invalidate } from '$app/navigation';

	let { data } = $props();

	let syncing = $state(false);

	/**
	 * Bring the banks up to date behind the arrival.
	 *
	 * The load function deliberately does not do this: it blocks navigation
	 * until it returns, and a first fill is unbounded pages across every
	 * account. The rows already held paint first with the date they are synced
	 * through, and this replaces them when the banks answer.
	 *
	 * `arrival` is bounded by the per-account interval, so reloading does not
	 * multiply the calls. Refresh is the member asking in as many words and is
	 * not bound by it.
	 */
	async function sync(action: 'sync-transactions' | 'refresh-transactions') {
		if (syncing) return;
		syncing = true;
		try {
			await fetch('/api/banking', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ action, accountId: data.accountId })
			});
			await invalidate('wimm:ledger');
		} finally {
			syncing = false;
		}
	}

	/**
	 * Once per arrival, and not once per render.
	 *
	 * `afterNavigate` fires on the initial load and on every navigation after
	 * it, and **not** on an invalidation — which is the whole point. As an
	 * `$effect` this was an endless loop: the effect read the syncing flag and
	 * the page data, the sync changed both, and each change re-ran the effect,
	 * which synced again. An effect that fires a request which invalidates its
	 * own dependency loops by construction, however it is guarded.
	 *
	 * Only the newest page asks. A cursor means the member is reading somewhere
	 * that new transactions do not belong.
	 */
	afterNavigate(() => {
		if (data.syncOnArrival) void sync('sync-transactions');
	});
</script>

<TransactionsScreen
	days={data.days}
	count={data.count}
	freshness={data.freshness}
	span={data.span}
	olderHref={data.olderHref}
	newerHref={data.newerHref}
	atOldest={data.atOldest}
	filterAccount={data.filterAccount}
	narrow={data.narrow}
	problems={data.problems}
	noBank={data.noBank}
	ownsNothing={data.ownsNothing}
	refreshing={syncing}
	onrefresh={() => sync('refresh-transactions')}
/>
