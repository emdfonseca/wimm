<script lang="ts" module>
	/** The filters in force, by the names they carry in the address. Empty is
	 *  not in force. */
	export interface LedgerFilterValues {
		account: string;
		q: string;
		/** `YYYY-MM`. */
		month: string;
		direction: '' | 'in' | 'out';
	}

	export interface FilterOption {
		value: string;
		/** `Savings · Monzo`, `August 2026`. */
		label: string;
	}

	export function filtersInForce(f: LedgerFilterValues): boolean {
		return f.account !== '' || f.q !== '' || f.month !== '' || f.direction !== '';
	}
</script>

<script lang="ts">
	import SearchField from '../atoms/SearchField.svelte';
	import SelectField from '../atoms/SelectField.svelte';
	import SegmentedControl from '../atoms/SegmentedControl.svelte';
	import Icon from '../atoms/Icon.svelte';
	import Button from '../atoms/Button.svelte';

	/**
	 * The ledger's filters: an account, a search, a month and a direction, and
	 * the one action that clears them.
	 *
	 * It is a GET form whose fields carry the address's own names, so without
	 * `onfilter` choosing something is a plain navigation and the search applies
	 * on Enter. With it, every change is reported once, with every other value as
	 * it was, and the caller turns it into an address; the search applies as the
	 * member types, and says so with `live`.
	 */
	interface Props {
		filters: LedgerFilterValues;
		/** Every account the member can narrow to, without `All accounts`. */
		accounts: FilterOption[];
		/** The months holding a match, newest first, without `All months`. */
		months: FilterOption[];
		/** Where Clear filters goes: the ledger with no filter. */
		clearHref: string;
		/** False where the screen offers Clear filters itself, so there is one
		 *  link with that name. */
		offerClear?: boolean;
		/** A phone: the search, and the other choices in a sheet. */
		compact?: boolean;
		/** `live` is true for a search applied by a pause in typing, which the
		 *  caller keeps out of the history after the first. */
		onfilter?: (next: LedgerFilterValues, how: { live: boolean }) => void;
		/** Clear filters was activated, before it navigates. */
		onclear?: () => void;
	}

	let {
		filters,
		accounts,
		months,
		clearHref,
		offerClear = true,
		compact = false,
		onfilter,
		onclear
	}: Props = $props();

	let form = $state<HTMLFormElement | null>(null);
	let sheet = $state<HTMLDialogElement | null>(null);

	const directions = [
		{ value: '', label: 'All' },
		{ value: 'in', label: 'Money in' },
		{ value: 'out', label: 'Money out' }
	];

	const accountOptions = $derived([{ value: '', label: 'All accounts' }, ...accounts]);
	const monthOptions = $derived([{ value: '', label: 'All months' }, ...months]);
	// A single month offered and none chosen is a control with nowhere to go.
	const showMonth = $derived(months.length > 1 || filters.month !== '');
	const inForce = $derived(filtersInForce(filters));
	/** The choices behind the Filters button, for its badge: not the search,
	 *  which stays in view. */
	const chosen = $derived(
		[filters.account, filters.month, filters.direction].filter((v) => v !== '').length
	);

	// What the direction's hidden field submits; the segmented control is a
	// radiogroup of buttons and carries no value of its own.
	let direction = $derived<string>(filters.direction);

	function apply(change: Partial<LedgerFilterValues>, live = false) {
		if (onfilter) {
			onfilter({ ...filters, ...change }, { live });
			return;
		}
		// A plain form submitting on every pause would fill the history.
		if (live) return;
		// Without a callback the form is the navigation. The hidden field has to
		// hold the new direction before it submits.
		if (change.direction !== undefined) direction = change.direction;
		queueMicrotask(() => form?.requestSubmit());
	}

	function onsubmit(event: SubmitEvent) {
		if (!onfilter) return;
		event.preventDefault();
	}
</script>

{#snippet controls(hideLabel: boolean)}
	<div class="account">
		<SelectField
			label="Account"
			name="account"
			{hideLabel}
			options={accountOptions}
			value={filters.account}
			onchange={(account) => apply({ account })}
		/>
	</div>

	{#if showMonth}
		<div class="month">
			<SelectField
				label="Month"
				name="month"
				{hideLabel}
				options={monthOptions}
				value={filters.month}
				onchange={(month) => apply({ month })}
			/>
		</div>
	{/if}

	<div class="direction">
		{#if !hideLabel}
			<span class="label" aria-hidden="true">Direction</span>
		{/if}
		<SegmentedControl
			label="Direction"
			options={directions}
			value={filters.direction}
			onchange={(next) => apply({ direction: next as LedgerFilterValues['direction'] })}
		/>
		<input type="hidden" name="direction" value={direction} />
	</div>
{/snippet}

<form
	bind:this={form}
	class="filters"
	class:compact
	role="search"
	aria-label="Filter transactions"
	method="get"
	action="/transactions"
	{onsubmit}
>
	<div class="search">
		<SearchField
			label="Search"
			name="q"
			hideLabel
			placeholder="Who it was with, or the bank's line"
			value={filters.q}
			onsearch={(q, { live }) => apply({ q }, live)}
		/>
	</div>

	{#if compact}
		<!-- On a phone the three choices wait in a sheet: the search is what a
		     member reaches for, and the list should start on the first screen. -->
		<button
			type="button"
			class="open-sheet"
			aria-haspopup="dialog"
			onclick={() => sheet?.showModal()}
		>
			<Icon name="sliders-horizontal" size={16} />
			Filters
			{#if chosen > 0}
				<span class="badge">{chosen}<span class="hidden"> chosen</span></span>
			{/if}
		</button>

		<dialog bind:this={sheet} class="sheet" aria-labelledby="ledger-filters-title">
			<h2 id="ledger-filters-title" class="sheet-title">Filters</h2>
			{@render controls(false)}
			<div class="sheet-actions">
				{#if inForce && offerClear}
					<a class="clear" href={clearHref} onclick={() => onclear?.()}>Clear filters</a>
				{/if}
				<Button onclick={() => sheet?.close()}>Done</Button>
			</div>
		</dialog>
	{:else}
		{@render controls(true)}
		{#if inForce && offerClear}
			<a class="clear" href={clearHref} onclick={() => onclear?.()}>Clear filters</a>
		{/if}
	{/if}
</form>

<style>
	.filters {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
		margin: 0;
		font-family: var(--type-family-body);
	}

	.search {
		flex: 1 1 260px;
		min-inline-size: 200px;
	}

	.account {
		flex: 0 1 220px;
		min-inline-size: 160px;
	}

	.month {
		flex: 0 1 170px;
		min-inline-size: 140px;
	}

	.direction {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.label {
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-sm);
	}

	.clear {
		display: inline-flex;
		align-items: center;
		block-size: var(--control-height-sm);
		padding-inline: var(--space-1);
		color: var(--color-accent);
		font-size: var(--type-size-body-sm);
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
	}

	.sheet-actions .clear {
		margin-inline-end: auto;
	}

	.clear:focus-visible,
	.open-sheet:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.compact {
		flex-wrap: nowrap;
	}

	.compact .search {
		flex: 1 1 auto;
		min-inline-size: 0;
	}

	.open-sheet {
		display: inline-flex;
		flex: none;
		align-items: center;
		gap: 6px;
		block-size: var(--control-height-md);
		padding-inline: 12px;
		border: 1px solid var(--color-control-border);
		border-radius: var(--radius-control);
		background: var(--color-bg-surface);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		cursor: pointer;
	}

	.badge {
		display: inline-grid;
		place-items: center;
		min-inline-size: 18px;
		block-size: 18px;
		padding-inline: 5px;
		border-radius: var(--radius-pill);
		background: var(--color-accent-subtle);
		color: var(--color-accent);
		font-size: 11px;
		font-weight: 600;
	}

	.sheet {
		position: fixed;
		inset-block: auto 0;
		inset-inline: 0;
		inline-size: 100%;
		max-inline-size: none;
		margin: 0;
		padding: var(--space-4) var(--space-4) var(--space-6);
		border: 0;
		border-radius: var(--radius-lg) var(--radius-lg) 0 0;
		background: var(--color-bg-surface);
		color: var(--color-text-primary);
	}

	.sheet[open] {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.sheet::backdrop {
		background: var(--color-scrim);
	}

	.sheet-title {
		margin: 0;
		font-size: var(--type-size-heading-sm);
		font-weight: 600;
	}

	.sheet-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
		margin-block-start: var(--space-2);
	}

	.hidden {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
</style>
