<script lang="ts" module>
	/**
	 * J03.A / 01 · Choose a bank.
	 *
	 * A full page rather than a dialog: connecting is complex, resumable and
	 * deep-linkable, which is exactly the case `docs/design/surfaces.md` names
	 * for a route-backed page.
	 */
	export type BankListState = 'default' | 'loading' | 'unavailable';

	export interface Bank {
		id: string;
		name: string;
		logoUrl?: string;
	}
</script>

<script lang="ts">
	import BankRow from '../molecules/BankRow.svelte';
	import ErrorNotice from '../molecules/ErrorNotice.svelte';
	import Page from '../templates/Page.svelte';
	import StepIndicator from '../molecules/StepIndicator.svelte';
	import StepActions from '../molecules/StepActions.svelte';

	interface Props {
		banks?: Bank[];
		state?: BankListState;
		onselect?: (bankId: string) => void;
		onretry?: () => void;
		/** Where "Back to Overview" goes. */
		backHref?: string;
		/** "Cancel", in the step's own footer — nothing has been chosen yet on
		 *  this step, so it is the same step back as the header's link. */
		oncancel?: () => void;
	}

	// Destructured under another name because a local variable called `state`
	// makes `$state(...)` below parse as a store subscription on it. The prop
	// stays `state`, as every other page in this library spells it.
	let {
		banks = [],
		state: listState = 'default',
		onselect,
		onretry,
		backHref = '/',
		oncancel
	}: Props = $props();

	/** Compact drops the header's back link and takes a shorter lede — both
	 *  drawn directly rather than passed down, so every caller does not need to
	 *  know a layout breakpoint that is this screen's concern alone. */
	let autoCompact = $state(false);

	$effect(() => {
		if (typeof window === 'undefined' || !window.matchMedia) return;
		const mq = window.matchMedia('(max-width: 767px)');
		autoCompact = mq.matches;
		const onchange = (e: MediaQueryListEvent) => (autoCompact = e.matches);
		mq.addEventListener('change', onchange);
		return () => mq.removeEventListener('change', onchange);
	});

	let query = $state('');

	const matches = $derived(
		query.trim() === ''
			? banks
			: banks.filter((bank) => bank.name.toLowerCase().includes(query.trim().toLowerCase()))
	);

	/**
	 * Announced in a live region, because a list narrowing under a member's
	 * fingers is a change they cannot see if they are not looking at it.
	 */
	const resultsMessage = $derived(
		query.trim() === ''
			? ''
			: matches.length === 0
				? 'No banks match'
				: `${matches.length} ${matches.length === 1 ? 'bank' : 'banks'} match`
	);
</script>

<Page title="Connect a bank" focusHeading>
	{#snippet action()}
		{#if !autoCompact}
			<a class="back" href={backHref}>Back to Overview</a>
		{/if}
	{/snippet}

	<StepIndicator current={1} total={2} />

	<p class="lede">
		{#if autoCompact}
			Pick your bank. You confirm at the bank itself.
		{:else}
			Pick the bank you want wimm to read. You confirm at the bank itself, and wimm never sees
			your banking password.
		{/if}
	</p>

	{#if listState === 'unavailable'}
		<ErrorNotice title="The bank list could not be loaded">
			wimm could not reach the service it uses to list banks. Nothing has been connected.
		</ErrorNotice>
		<button type="button" class="retry" onclick={onretry}>Try again</button>
	{:else}
		<label class="search">
			<span class="search-label">Search banks</span>
			<input
				type="search"
				bind:value={query}
				placeholder="Start typing a bank's name"
				autocomplete="off"
			/>
		</label>

		<!-- The count, not the list, is what is announced: reading every row
		     aloud on each keystroke is noise. -->
		<p class="results" role="status" aria-live="polite">{resultsMessage}</p>

		{#if listState === 'loading'}
			<p class="empty">Loading banks…</p>
		{:else if matches.length === 0}
			<div class="no-match">
				<p class="empty">No bank matches “{query}”.</p>
				<p class="empty-note">
					wimm can only connect banks its open-banking service covers. If yours is not here, it
					cannot be connected yet.
				</p>
			</div>
		{:else}
			<ul class="banks">
				{#each matches as bank (bank.id)}
					<li>
						<BankRow name={bank.name} logoUrl={bank.logoUrl} onselect={() => onselect?.(bank.id)} />
					</li>
				{/each}
			</ul>
		{/if}
	{/if}

	<StepActions lesserLabel="Cancel" onLesser={oncancel} />
</Page>

<style>
	.back {
		color: var(--color-accent);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 500;
		text-decoration: none;
	}

	.back:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.lede,
	.empty,
	.empty-note,
	.results {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.results,
	.empty-note {
		font-size: var(--type-size-body-sm);
	}

	.search {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.search-label {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	.search input {
		block-size: 36px;
		padding-inline: 12px;
		border: 1px solid var(--color-control-border);
		border-radius: var(--radius-control);
		background: var(--color-bg-surface);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}
	.search input:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.banks {
		display: flex;
		flex-direction: column;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.no-match {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.retry {
		border: none;
		background: none;
		padding: 0;
		color: var(--color-accent);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		text-decoration: underline;
		cursor: pointer;
	}
</style>
