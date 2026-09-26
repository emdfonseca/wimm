<script lang="ts">
	import { TransactionsScreen, type LedgerFilterValues } from '@wimm/ui';
	import { afterNavigate, goto, invalidate } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { filterQuery } from './filters';

	let { data } = $props();

	let syncing = $state(false);

	/**
	 * The current history entry came from typing a search. The next pause
	 * replaces it rather than adding another, so Back goes to before the
	 * search instead of stepping through half-typed words. Any navigation this
	 * page did not make — a link, Back — ends the search's run.
	 */
	let typing = false;

	afterNavigate(({ type }) => {
		if (type !== 'goto') typing = false;
	});

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
		// A pause in typing is not an arrival: the search's first result already
		// was, and asking again every few letters would re-read the page twice.
		if (data.syncOnArrival && !typing) void sync('sync-transactions');
	});

	/**
	 * A filter change is a navigation, so Back undoes it. Focus stays on the
	 * control the member just used, and the screen announces the outcome.
	 */
	function onfilter(next: LedgerFilterValues, { live }: { live: boolean }) {
		const replaceState = live && typing;
		typing = live;
		// resolve() takes a route, not a query string.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		void goto(`${resolve('/(app)/transactions')}${filterQuery(next)}`, {
			keepFocus: true,
			noScroll: true,
			replaceState
		});
	}
</script>

<TransactionsScreen
	days={data.days}
	count={data.count}
	freshness={data.freshness}
	span={data.span}
	atOldest={data.atOldest}
	pages={data.pages}
	currentPage={data.currentPage}
	newestHref={data.newestHref}
	oldestHref={data.oldestHref}
	filters={data.filters}
	accounts={data.accounts}
	months={data.months}
	clearHref={data.clearHref}
	totals={data.totals}
	transfersLeftOut={data.transfersLeftOut}
	notSettled={data.notSettled}
	{onfilter}
	narrow={data.narrow}
	problems={data.problems}
	noBank={data.noBank}
	ownsNothing={data.ownsNothing}
	refreshing={syncing}
	onrefresh={() => sync('refresh-transactions')}
/>
