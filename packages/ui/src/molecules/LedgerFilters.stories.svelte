<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
	import LedgerFilters from './LedgerFilters.svelte';

	const accounts = [
		{ value: 'acc-current', label: 'Current account · Monzo' },
		{ value: 'acc-savings', label: 'Savings · Monzo' },
		{ value: 'acc-casa', label: 'Casa CC · Montepio' }
	];

	const months = [
		{ value: '2026-09', label: 'September 2026' },
		{ value: '2026-08', label: 'August 2026' },
		{ value: '2026-07', label: 'July 2026' }
	];

	const none = { account: '', q: '', month: '', direction: '' as const };
	const everything = {
		account: 'acc-current',
		q: 'galp',
		month: '2026-08',
		direction: 'out' as const
	};

	const { Story } = defineMeta({
		title: 'Molecules/LedgerFilters',
		component: LedgerFilters,
		tags: ['autodocs'],
		args: {
			filters: none,
			accounts,
			months,
			clearHref: '/transactions',
			onfilter: fn(),
			onclear: fn()
		}
	});
</script>

<!-- One landmark, reached in the order it reads, with nothing to clear. Each
     change is reported once, carrying every other value as it was; Enter
     applies the search without waiting for a pause. -->
<Story
	name="NothingChosen"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('search', { name: 'Filter transactions' })).toBeInTheDocument();
		await expect(canvas.queryByRole('link', { name: 'Clear filters' })).toBeNull();

		const account = canvas.getByRole('combobox', { name: 'Account' });
		const search = canvas.getByRole('searchbox', { name: 'Search' });
		const month = canvas.getByRole('combobox', { name: 'Month' });
		const all = canvas.getByRole('radio', { name: 'All' });
		await expect(account).toHaveValue('');
		await expect(month).toHaveValue('');
		await expect(all).toHaveAttribute('aria-checked', 'true');

		// The search leads: it is what a member reaches for first.
		search.focus();
		await userEvent.tab();
		await expect(account).toHaveFocus();
		await userEvent.tab();
		await expect(month).toHaveFocus();
		await userEvent.tab();
		await expect(all).toHaveFocus();

		await userEvent.selectOptions(account, 'acc-savings');
		await expect(args.onfilter).toHaveBeenCalledTimes(1);
		await expect(args.onfilter).toHaveBeenLastCalledWith(
			{ ...none, account: 'acc-savings' },
			{ live: false }
		);

		await userEvent.type(search, ' galp ');
		await expect(args.onfilter).toHaveBeenCalledTimes(1);
		await userEvent.keyboard('{Enter}');
		await expect(args.onfilter).toHaveBeenCalledTimes(2);
		await expect(args.onfilter).toHaveBeenLastCalledWith({ ...none, q: 'galp' }, { live: false });

		await userEvent.selectOptions(month, '2026-08');
		await expect(args.onfilter).toHaveBeenCalledTimes(3);
		await expect(args.onfilter).toHaveBeenLastCalledWith(
			{ ...none, month: '2026-08' },
			{ live: false }
		);

		await userEvent.click(canvas.getByRole('radio', { name: 'Money in' }));
		await expect(args.onfilter).toHaveBeenCalledTimes(4);
		await expect(args.onfilter).toHaveBeenLastCalledWith(
			{ ...none, direction: 'in' },
			{ live: false }
		);
	}}
/>

<!-- Everything chosen, so the search can be cleared and so can every filter.
     The tab order takes in both. -->
<Story
	name="EverythingChosen"
	args={{ filters: everything }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		const account = canvas.getByRole('combobox', { name: 'Account' });
		const search = canvas.getByRole('searchbox', { name: 'Search' });
		const clearSearch = canvas.getByRole('button', { name: 'Clear search' });
		const month = canvas.getByRole('combobox', { name: 'Month' });
		const out = canvas.getByRole('radio', { name: 'Money out' });
		const clear = canvas.getByRole('link', { name: 'Clear filters' });

		await expect(account).toHaveValue('acc-current');
		await expect(search).toHaveValue('galp');
		await expect(month).toHaveValue('2026-08');
		await expect(out).toHaveAttribute('aria-checked', 'true');
		await expect(clear).toHaveAttribute('href', '/transactions');

		search.focus();
		for (const next of [clearSearch, account, month, out, clear]) {
			await userEvent.tab();
			await expect(next).toHaveFocus();
		}

		await userEvent.click(clearSearch);
		await expect(args.onfilter).toHaveBeenCalledTimes(1);
		await expect(args.onfilter).toHaveBeenLastCalledWith({ ...everything, q: '' }, { live: false });
		await expect(search).toHaveFocus();
	}}
/>

<!-- One month holding everything and none chosen: there is nowhere to go, so
     there is no Month select. -->
<Story
	name="OneMonthOnly"
	args={{ months: [months[0]!] }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByRole('combobox', { name: 'Month' })).toBeNull();
		await expect(canvas.getByRole('combobox', { name: 'Account' })).toBeInTheDocument();
	}}
/>

<Story
	name="Compact"
	args={{ filters: { ...none, q: 'galp', month: '2026-08' }, compact: true }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		// One row: the search and a button naming how many choices are made.
		await expect(canvas.getByRole('searchbox', { name: 'Search' })).toHaveValue('galp');
		await expect(canvas.queryByRole('combobox')).toBeNull();
		const open = canvas.getByRole('button', { name: 'Filters 1 chosen' });
		await expect(open).toHaveAttribute('aria-haspopup', 'dialog');

		// The sheet holds the choices, labelled, in the order they read, and
		// Done hands focus back to the button that opened it.
		await userEvent.click(open);
		const sheet = within(canvas.getByRole('dialog', { name: 'Filters' }));
		const order = [
			sheet.getByRole('combobox', { name: 'Account' }),
			sheet.getByRole('combobox', { name: 'Month' }),
			sheet.getByRole('radiogroup', { name: 'Direction' }),
			sheet.getByRole('link', { name: 'Clear filters' })
		];
		await expect(sheet.getByText('Account')).toBeVisible();
		const tops = order.map((el) => el.getBoundingClientRect().top);
		for (let i = 1; i < tops.length; i++) {
			await expect(tops[i]!).toBeGreaterThan(tops[i - 1]!);
		}
		await userEvent.selectOptions(order[1] as HTMLSelectElement, '2026-07');
		await expect(args.onfilter).toHaveBeenLastCalledWith(
			{ ...none, q: 'galp', month: '2026-07' },
			{ live: false }
		);
		await userEvent.click(sheet.getByRole('button', { name: 'Done' }));
		await expect(canvas.queryByRole('dialog')).toBeNull();
		await expect(open).toHaveFocus();
	}}
/>

<!-- Without a callback the bar is a plain GET form whose fields carry the
     address's own names. -->
<Story
	name="AsAPlainForm"
	args={{ filters: everything, onfilter: undefined }}
	play={async ({ canvasElement }) => {
		const form = within(canvasElement).getByRole('search', { name: 'Filter transactions' });
		await expect(form).toHaveAttribute('method', 'get');
		await expect(form).toHaveAttribute('action', '/transactions');
		const names = [...(form as HTMLFormElement).elements]
			.map((el) => el.getAttribute('name'))
			.filter(Boolean);
		await expect(names).toEqual(['q', 'account', 'month', 'direction']);
		await expect(new FormData(form as HTMLFormElement).get('direction')).toBe('out');
	}}
/>

<!-- The search applies as the member types. The change is marked live, so the
     route can keep half-typed words out of the history. -->
<Story
	name="SearchingAsYouType"
	play={async ({ canvasElement, args }) => {
		const search = within(canvasElement).getByRole('searchbox', { name: 'Search' });
		await userEvent.type(search, 'galp');
		await waitFor(() => expect(args.onfilter).toHaveBeenCalledTimes(1));
		await expect(args.onfilter).toHaveBeenLastCalledWith({ ...none, q: 'galp' }, { live: true });
	}}
/>
