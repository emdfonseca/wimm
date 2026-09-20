<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import LedgerRow from './LedgerRow.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/LedgerRow',
		component: LedgerRow,
		tags: ['autodocs'],
		// A row fills the list it is in. Centred, it shrinks to fit and its
		// columns squeeze the name to "S…", which no list ever shows.
		parameters: { layout: 'padded' },
		args: {
			description: 'Pingo Doce',
			account: 'Current account',
			amount: '−€42.18',
			date: '17 Sep',
			negative: true
		}
	});
</script>

<Story name="MoneyLeaving" />

<!-- The sign is what tells the two directions apart, so the assertion is on
     the character rather than on a colour. -->
<Story
	name="MoneyArriving"
	args={{ description: 'Salary', amount: '+€2,180.00', negative: false }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('+€2,180.00')).toBeInTheDocument();
	}}
/>

<!-- Money that has moved as far as the member is concerned. The badge carries
     text beside its colour, in both themes. -->
<Story
	name="NotSettled"
	args={{
		description: 'Transfer to Ana Reis',
		account: 'Joint savings',
		amount: '−€60.00',
		unsettled: true
	}}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('Not settled')).toBeInTheDocument();
	}}
/>

<!-- A bank that named nobody: wimm shows what it did give and never invents. -->
<Story
	name="TheBankNamedNobody"
	args={{ description: 'Card payment', amount: '−€4.50', initials: '—' }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('Card payment')).toBeInTheDocument();
	}}
/>

<!-- There is no category column and no selection control: this list is read,
     not worked. -->
<Story
	name="NoCategoryAndNoSelection"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByRole('checkbox')).not.toBeInTheDocument();
		await expect(canvas.queryByText(/Category/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="Compact"
	args={{ compact: true }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Pingo Doce')).toBeInTheDocument();
		await expect(canvas.getByText('Current account')).toBeInTheDocument();
	}}
/>

<Story
	name="CompactNotSettled"
	args={{ compact: true, unsettled: true, description: 'Transfer to Ana Reis' }}
	globals={{ viewport: { value: 'compact' } }}
/>

<!-- A description with nowhere to go still leaves the amount readable. -->
<Story
	name="LongContent"
	args={{
		description: 'Standing order to the management company for the building maintenance fund',
		account: 'Joint savings'
	}}
/>

<!-- The name is wimm's; the line under it is the bank's, exactly as written,
     so a derived name can be checked against the statement. -->
<Story
	name="BanksLineShown"
	args={{ banksLine: 'COMPRA PINGO DOCE LISBOA 230002268342127', hideDate: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Pingo Doce')).toBeInTheDocument();
		await expect(canvas.getByText('COMPRA PINGO DOCE LISBOA 230002268342127')).toBeInTheDocument();
		await expect(canvasElement.querySelector('.ledger-row')!.getBoundingClientRect().height).toBe(
			56
		);
	}}
/>

<!-- Where the bank's line is the name, saying it twice is noise. -->
<Story
	name="BanksLineSameAsName"
	args={{
		description: 'Salary',
		amount: '+€2,180.00',
		negative: false,
		banksLine: 'Salary',
		hideDate: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getAllByText('Salary')).toHaveLength(1);
		await expect(canvasElement.querySelector('.ledger-row')!.getBoundingClientRect().height).toBe(
			56
		);
	}}
/>

<Story
	name="CompactBanksLineShown"
	args={{ compact: true, banksLine: 'COMPRA PINGO DOCE LISBOA 230002268342127' }}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Pingo Doce')).toBeInTheDocument();
		await expect(canvas.getByText('COMPRA PINGO DOCE LISBOA 230002268342127')).toBeInTheDocument();
		await expect(canvasElement.querySelector('.ledger-row')!.getBoundingClientRect().height).toBe(
			80
		);
	}}
/>

<Story name="AList" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; flex-direction: column; inline-size: 880px;">
			<LedgerRow {...args} />
			<LedgerRow
				{...args}
				description="Transfer to Ana Reis"
				account="Joint savings"
				amount="−€60.00"
				unsettled
			/>
			<LedgerRow
				{...args}
				description="Salary"
				amount="+€2,180.00"
				negative={false}
				date="16 Sep"
			/>
		</div>
	{/snippet}
</Story>
