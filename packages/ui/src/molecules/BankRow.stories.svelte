<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import BankRow from './BankRow.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/BankRow',
		component: BankRow,
		tags: ['autodocs'],
		argTypes: { disabled: { control: 'boolean' } },
		args: { name: 'Caixa Económica Montepio Geral', onselect: fn() }
	});

	// A wide wordmark, deliberately not square — the shape the logo fix has
	// to letterbox rather than crop.
	const wideLogo =
		'data:image/svg+xml,' +
		encodeURIComponent(
			'<svg xmlns="http://www.w3.org/2000/svg" width="240" height="48"><rect width="240" height="48" fill="steelblue"/></svg>'
		);
</script>

<Story name="Rest" />

<Story name="A short name" args={{ name: 'Revolut' }} />

<!-- A bank whose name is longer than the row truncates rather than wrapping:
     the row's height is shared with the account row beside it. -->
<Story
	name="A name longer than the row"
	args={{ name: 'Banco Comercial Português — ActivoBank Particulares' }}
/>

<!-- It is a button, not a link: picking a bank begins a hand-off out of the
     product, and a link would offer a new tab where the return cannot land. -->
<Story
	name="It is a button"
	play={async ({ canvasElement }) => {
		const row = within(canvasElement).getByRole('button');
		await expect(row).toBeInTheDocument();
		await expect(row.tagName).toBe('BUTTON');
	}}
/>

<Story
	name="Selecting"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button'));
		await expect(args.onselect).toHaveBeenCalledOnce();
	}}
/>

<!-- Reachable and operable from the keyboard alone, which every step of
     connecting a bank has to be. -->
<Story
	name="Keyboard"
	play={async ({ canvasElement, args }) => {
		await userEvent.tab();
		await expect(within(canvasElement).getByRole('button')).toHaveFocus();
		await userEvent.keyboard('{Enter}');
		await expect(args.onselect).toHaveBeenCalled();
	}}
/>

<Story
	name="Disabled"
	args={{ disabled: true }}
	play={async ({ canvasElement, args }) => {
		const row = within(canvasElement).getByRole('button');
		await expect(row).toBeDisabled();
		await userEvent.click(row, { pointerEventsCheck: 0 });
		await expect(args.onselect).not.toHaveBeenCalled();
	}}
/>

<!-- The accessible name is the bank's name, and the mark adds nothing to it:
     initials read aloud beside the name would be noise. -->
<Story
	name="Named by the bank"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('button')).toHaveAccessibleName(
			'Caixa Económica Montepio Geral'
		);
	}}
/>

<Story name="A list of banks" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; flex-direction: column; inline-size: 360px;">
			<BankRow {...args} name="Caixa Económica Montepio Geral" />
			<BankRow {...args} name="ActivoBank" />
			<BankRow {...args} name="Revolut" />
		</div>
	{/snippet}
</Story>

<!-- A wide wordmark beside a square logo: neither is cropped, and both names
     start at the same horizontal position (specs/banking/bank-connections:
     "A bank's logo renders whole, whatever its shape"). -->
<Story name="A wide logo next to a square one" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; flex-direction: column; inline-size: 360px;">
			<BankRow {...args} name="Wide Wordmark Bank" logoUrl={wideLogo} />
			<BankRow {...args} name="Square Bank" />
		</div>
	{/snippet}
</Story>

<Story
	name="Neither logo is cropped, and names stay aligned"
	tags={['!autodocs']}
	play={async ({ canvasElement }) => {
		const marks = canvasElement.querySelectorAll<HTMLElement>('.mark');
		const names = canvasElement.querySelectorAll<HTMLElement>('.name');
		await expect(marks).toHaveLength(2);
		for (const mark of marks) {
			await expect(getComputedStyle(mark).objectFit).toBe('contain');
		}
		const [firstLeft, secondLeft] = Array.from(names).map((n) => n.getBoundingClientRect().left);
		await expect(firstLeft).toBe(secondLeft);
	}}
>
	{#snippet template(args)}
		<div style="display: flex; flex-direction: column; inline-size: 360px;">
			<BankRow {...args} name="Wide Wordmark Bank" logoUrl={wideLogo} />
			<BankRow {...args} name="Square Bank" />
		</div>
	{/snippet}
</Story>

