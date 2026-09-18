<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import AccountRow from './AccountRow.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/AccountRow',
		component: AccountRow,
		tags: ['autodocs'],
		args: {
			name: 'Conta à Ordem',
			bank: 'Montepio',
			numberSuffix: '0538',
			balance: '€4,200.10',
			readAt: '2 minutes ago'
		}
	});
</script>

<Story name="Fresh" />

<!-- A bank that could not be reached keeps its figure and its original time.
     Losing the previous number is the one thing that must not happen. -->
<Story
	name="Stale"
	args={{ stale: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('€4,200.10')).toBeInTheDocument();
		await expect(canvas.getByText(/Last read 2 minutes ago/)).toBeInTheDocument();
		await expect(canvas.getByText('Could not update')).toBeInTheDocument();
	}}
/>

<!-- Access has run out: different from a bank having a bad minute, and it
     routes the member to restoring rather than retrying. -->
<Story
	name="No longer updating"
	args={{ notUpdating: true }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('Not updating')).toBeInTheDocument();
	}}
/>

<!-- Direction never rests on colour: the figure carries its own sign. -->
<Story
	name="Overdrawn"
	args={{ balance: '−€312.40', negative: true }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText(/−/)).toBeInTheDocument();
	}}
/>

<Story name="A zero balance" args={{ balance: '€0.00' }} />

<!-- At balance level the server sends no identifier, so none is rendered and
     none is invented. -->
<Story
	name="Seen at balance level"
	args={{ numberSuffix: undefined }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Montepio')).toBeInTheDocument();
		await expect(canvas.queryByText(/••••/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="Seen in full"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText(/•••• 0538/)).toBeInTheDocument();
	}}
/>

<!-- A bank that gave no product name still renders as something. -->
<Story name="An account the bank gave no name" args={{ name: 'Account', numberSuffix: '7712' }} />

<!-- A balance is never shown without the time it was read. -->
<Story
	name="No reading yet"
	args={{ balance: undefined, readAt: undefined }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).queryByText(/€/)).not.toBeInTheDocument();
	}}
/>

<!-- Left out: the figures never render, whatever balance was passed, because a
     left-out account is never read. -->
<Story
	name="Left out"
	args={{ leftOut: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Left out')).toBeInTheDocument();
		await expect(canvas.queryByText('€4,200.10')).not.toBeInTheDocument();
	}}
/>

<Story name="A list" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; flex-direction: column; inline-size: 520px;">
			<AccountRow {...args} />
			<AccountRow {...args} name="Poupança" numberSuffix="7712" balance="€11,930.00" />
			<AccountRow
				{...args}
				name="Conta ActivoBank"
				bank="ActivoBank"
				numberSuffix="5594"
				balance="−€312.40"
				negative
				stale
			/>
		</div>
	{/snippet}
</Story>
