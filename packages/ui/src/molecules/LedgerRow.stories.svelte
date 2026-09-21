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

<!-- The label is text in the status slot and the row keeps its height. -->
<Story
	name="Unusual"
	args={{ description: 'Galp', account: 'Current account', amount: '−€210.00', unusual: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Galp')).toBeInTheDocument();
		await expect(canvas.getByText('−€210.00')).toBeInTheDocument();
		await expect(canvas.getByText('Unusual')).toBeInTheDocument();
		await expect(canvas.queryByText('Unusual income')).not.toBeInTheDocument();
		await expect(canvasElement.querySelector('.ledger-row')!.getBoundingClientRect().height).toBe(
			56
		);
	}}
/>

<Story
	name="CompactUnusual"
	args={{
		compact: true,
		description: 'Galp',
		account: 'Current account',
		amount: '−€210.00',
		unusual: true
	}}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Galp')).toBeInTheDocument();
		await expect(canvas.getByText('−€210.00')).toBeInTheDocument();
		await expect(canvas.getByText('Unusual')).toBeInTheDocument();
	}}
/>

<!-- The sign says which way it went and so does the label. -->
<Story
	name="UnusualIncome"
	args={{
		description: 'Employer Lda',
		account: 'Current account',
		amount: '+€9,804.00',
		negative: false,
		unusual: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Employer Lda')).toBeInTheDocument();
		await expect(canvas.getByText('+€9,804.00')).toBeInTheDocument();
		await expect(canvas.getByText('Unusual income')).toBeInTheDocument();
		await expect(canvas.queryByText('Unusual')).not.toBeInTheDocument();
	}}
/>

<!-- A recurring payment: the cadence and account in the account's place, the
     expected date in the date's, no settled marker. -->
<Story
	name="ExpectedDate"
	args={{
		description: 'Spotify',
		account: 'Monthly · Current account · Monzo',
		amount: '−€9.99',
		date: 'Was expected 18 Sep',
		hideStatus: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Spotify')).toBeInTheDocument();
		await expect(canvas.getByText('Monthly · Current account · Monzo')).toBeInTheDocument();
		const date = canvas.getByText('Was expected 18 Sep');
		await expect(date).toBeInTheDocument();
		await expect(date.scrollWidth).toBeLessThanOrEqual(date.clientWidth);
		await expect(canvas.getByText('−€9.99')).toBeInTheDocument();
		await expect(canvasElement.querySelector('.ledger-row')!.getBoundingClientRect().height).toBe(
			56
		);
	}}
/>

<!-- One link holds everything the row says, so a screen reader reads a single
     destination rather than four stops. -->
<Story
	name="WithANoteAndALink"
	args={{
		description: 'Galp',
		account: 'Current account',
		amount: '−€210.00',
		date: '27 Mar',
		unusual: true,
		note: 'usually about €60',
		href: '/transactions?page=2026-03-27.abc'
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const links = canvas.getAllByRole('link');
		await expect(links).toHaveLength(1);
		const link = links[0]!;
		for (const words of ['Galp', '27 Mar', '−€210.00', 'usually about €60']) {
			await expect(link).toHaveTextContent(words);
		}
		await expect(link).toHaveAttribute('href', '/transactions?page=2026-03-27.abc');
		await expect(canvasElement.querySelector('.ledger-row')!.getBoundingClientRect().height).toBe(
			56
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

<!-- Half of a movement between two accounts the member owns. The label is text
     in the same pill `Unusual` uses, never colour and never an icon. -->
<Story
	name="BetweenYourAccounts"
	args={{
		description: 'Transfer to savings',
		account: 'Current account',
		amount: '−€500.00',
		date: '1 Sep',
		transfer: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Transfer to savings')).toBeInTheDocument();
		await expect(canvas.getByText('Current account')).toBeInTheDocument();
		await expect(canvas.getByText('1 Sep')).toBeInTheDocument();
		await expect(canvas.getByText('−€500.00')).toBeInTheDocument();
		await expect(canvas.getByText('Between your accounts')).toBeInTheDocument();
	}}
/>

<Story
	name="CompactBetweenYourAccounts"
	args={{
		description: 'Transfer to savings',
		account: 'Current account',
		amount: '−€500.00',
		date: '1 Sep',
		transfer: true,
		compact: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Transfer to savings')).toBeInTheDocument();
		await expect(canvas.getByText('Current account')).toBeInTheDocument();
		await expect(canvas.getByText('1 Sep')).toBeInTheDocument();
		await expect(canvas.getByText('−€500.00')).toBeInTheDocument();
		await expect(canvas.getByText('Between your accounts')).toBeInTheDocument();
	}}
/>

<!-- The arriving half reads the same way. Nothing on it names the account the
     money came from. -->
<Story
	name="BetweenYourAccountsArriving"
	args={{
		description: 'Transfer from current account',
		account: 'Savings',
		amount: '+€500.00',
		date: '1 Sep',
		negative: false,
		transfer: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Transfer from current account')).toBeInTheDocument();
		await expect(canvas.getByText('Savings')).toBeInTheDocument();
		await expect(canvas.getByText('1 Sep')).toBeInTheDocument();
		await expect(canvas.getByText('+€500.00')).toBeInTheDocument();
		await expect(canvas.getByText('Between your accounts')).toBeInTheDocument();
	}}
/>

<!-- The slot holds one status. A row told it is both shows the transfer, which
     is the stronger statement: a transfer is never also unusual. -->
<Story
	name="ATransferIsNeverAlsoUnusual"
	args={{
		description: 'Transfer to savings',
		account: 'Current account',
		amount: '−€6,000.00',
		date: '1 Sep',
		transfer: true,
		unusual: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Between your accounts')).toBeInTheDocument();
		await expect(canvas.queryByText('Unusual')).not.toBeInTheDocument();
	}}
/>

<!-- Not settled wins over both: whether the money has moved outranks what kind
     of movement it is. -->
<Story
	name="NotSettledWinsOverATransfer"
	args={{
		description: 'Transfer to savings',
		account: 'Current account',
		amount: '−€500.00',
		date: '1 Sep',
		transfer: true,
		unsettled: true
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Not settled')).toBeInTheDocument();
		await expect(canvas.queryByText('Between your accounts')).not.toBeInTheDocument();
	}}
/>
