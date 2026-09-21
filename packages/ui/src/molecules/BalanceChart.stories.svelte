<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, waitFor, within } from 'storybook/test';
	import BalanceChart from './BalanceChart.svelte';
	import { balanceChart, balancePoints, balanceSummary } from './BalanceChart.fixture';

	const index = (date: string) => balancePoints.findIndex((p) => p.date === date);
	const live = (root: Element) => root.querySelector('[aria-live="polite"]')!;
	const popover = (root: Element) => root.querySelector<HTMLElement>('.popover');

	const { Story } = defineMeta({
		title: 'Molecules/BalanceChart',
		component: BalanceChart,
		tags: ['autodocs'],
		args: balanceChart
	});
</script>

<!-- The ends and the extremes are readable without touching it, and the
     chart is one tab stop that a keyboard can walk. -->
<Story
	name="AsItOpens"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Balance · EUR')).toBeInTheDocument();
		await expect(canvas.getByText('22 Jun to 20 Sep')).toBeInTheDocument();
		await expect(canvas.getByText('€12,118')).toBeInTheDocument();
		await expect(canvas.getByText('€9,204')).toBeInTheDocument();
		await expect(canvas.getByText('22 Jun')).toBeInTheDocument();
		await expect(canvas.getByText('20 Sep')).toBeInTheDocument();

		const plot = canvas.getByRole('img', { name: balanceSummary });
		await userEvent.tab();
		await expect(plot).toHaveFocus();
		await expect(live(canvasElement).textContent).toBe('');
		await expect(canvasElement.querySelector('.popover')).toBeNull();
	}}
/>

<!-- The popover is the readout made visible, so the live region says what it
     shows, all of it, on every key. -->
<Story
	name="OpensFromTheKeyboard"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.tab();
		await expect(canvas.getByRole('img', { name: balanceSummary })).toHaveFocus();

		await userEvent.keyboard('{ArrowLeft}');
		await expect(live(canvasElement)).toHaveTextContent(
			'19 Sep. €11,717.13. €23.58 less than 18 Sep. No transactions this day.'
		);
		await expect(popover(canvasElement)).toHaveTextContent('19 Sep');
		await userEvent.keyboard('{Home}');
		await expect(live(canvasElement)).toHaveTextContent(
			'22 Jun. €9,640.00. No transactions this day.'
		);
		await expect(popover(canvasElement)).toHaveTextContent('No transactions this day.');
		await userEvent.keyboard('{End}');
		await expect(live(canvasElement)).toHaveTextContent(
			'20 Sep. €11,693.55. €23.58 less than 19 Sep. No transactions this day.'
		);
		await expect(popover(canvasElement)).toHaveTextContent('€11,693.55');
		await userEvent.keyboard('{Escape}');
		await expect(popover(canvasElement)).toBeNull();
		await expect(live(canvasElement).textContent).toBe('');
	}}
/>

<!-- Touch has no hover and no arrow keys: a tap opens it, and it stays until
     the next tap, which is anywhere else. -->
<Story
	name="OpensFromATap"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const plot = canvas.getByRole('img', { name: balanceSummary });
		const box = plot.getBoundingClientRect();
		const tap = (target: EventTarget, clientX: number) =>
			target.dispatchEvent(
				new PointerEvent('pointerdown', {
					pointerType: 'touch',
					clientX,
					clientY: box.top + 10,
					bubbles: true
				})
			);

		tap(plot, box.left + 1);
		await waitFor(() =>
			expect(live(canvasElement)).toHaveTextContent('22 Jun. €9,640.00. No transactions this day.')
		);
		await expect(popover(canvasElement)).not.toBeNull();

		plot.dispatchEvent(new PointerEvent('pointerleave', { pointerType: 'touch' }));
		await new Promise((r) => setTimeout(r, 50));
		await expect(popover(canvasElement)).not.toBeNull();

		tap(canvasElement.ownerDocument.body, 0);
		await waitFor(() => expect(popover(canvasElement)).toBeNull());
		await expect(live(canvasElement).textContent).toBe('');
	}}
/>

<!-- What happened on the day it dropped, and which of it was unusual. -->
<Story
	name="ADayPointedAt"
	args={{ marked: index('3 Aug') }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('img', { name: balanceSummary })).toBeInTheDocument();
		const pop = within(popover(canvasElement)!);
		for (const words of [
			'3 Aug',
			'€9,870.12',
			'€805.79 less than 2 Aug',
			'Leroy Merlin',
			'−€640.00',
			'Unusual',
			'Galp',
			'−€92.10',
			'and 3 smaller'
		]) {
			await expect(pop.getByText(words)).toBeInTheDocument();
		}
		await expect(canvas.queryByText('3 Aug · €9,870.12')).not.toBeInTheDocument();
		await expect(live(canvasElement)).toHaveTextContent(
			'3 Aug. €9,870.12. €805.79 less than 2 Aug. Leroy Merlin −€640.00, Unusual. Galp −€92.10. And 3 smaller.'
		);
	}}
/>

<Story
	name="ADayWithNoTransactions"
	args={{ marked: index('9 Aug') }}
	play={async ({ canvasElement }) => {
		const pop = within(popover(canvasElement)!);
		for (const words of [
			'9 Aug',
			'€9,064.33',
			'No change from 8 Aug',
			'No transactions this day.'
		]) {
			await expect(pop.getByText(words)).toBeInTheDocument();
		}
		await expect(live(canvasElement)).toHaveTextContent(
			'9 Aug. €9,064.33. No change from 8 Aug. No transactions this day.'
		);
	}}
/>

<Story
	name="MoneyArriving"
	args={{ marked: index('15 Sep') }}
	play={async ({ canvasElement }) => {
		const pop = within(popover(canvasElement)!);
		for (const words of [
			'15 Sep',
			'€11,748.72',
			'€2,437.01 more than 14 Sep',
			'Salary',
			'+€2,450.00',
			'and 1 smaller'
		]) {
			await expect(pop.getByText(words)).toBeInTheDocument();
		}
		await expect(live(canvasElement)).toHaveTextContent(
			'15 Sep. €11,748.72. €2,437.01 more than 14 Sep. Salary +€2,450.00. And 1 smaller.'
		);
	}}
/>

<Story
	name="UnusualIncomeArriving"
	args={{ marked: index('28 Aug') }}
	play={async ({ canvasElement }) => {
		const pop = within(popover(canvasElement)!);
		for (const words of [
			'28 Aug',
			'€12,640.18',
			'€9,804.00 more than 27 Aug',
			'Employer Lda',
			'+€9,804.00',
			'Unusual income'
		]) {
			await expect(pop.getByText(words)).toBeInTheDocument();
		}
		await expect(pop.queryByText('Unusual')).not.toBeInTheDocument();
		await expect(live(canvasElement)).toHaveTextContent(
			'28 Aug. €12,640.18. €9,804.00 more than 27 Aug. Employer Lda +€9,804.00, Unusual income.'
		);
	}}
/>

<!-- No day before it, so no line naming one. -->
<Story
	name="TheFirstDay"
	args={{ marked: 0 }}
	play={async ({ canvasElement }) => {
		const pop = within(popover(canvasElement)!);
		await expect(pop.getByText('22 Jun')).toBeInTheDocument();
		await expect(pop.getByText('€9,640.00')).toBeInTheDocument();
		await expect(pop.getByText('No transactions this day.')).toBeInTheDocument();
		await expect(popover(canvasElement)).not.toHaveTextContent(/than|No change/);
		await expect(live(canvasElement)).toHaveTextContent(
			'22 Jun. €9,640.00. No transactions this day.'
		);
	}}
/>

<!-- Five rows and the count of the rest, never twelve. -->
<Story
	name="ABusyDay"
	args={{ marked: index('10 Sep') }}
	play={async ({ canvasElement }) => {
		const pop = popover(canvasElement)!;
		await expect(pop.querySelectorAll('.mover')).toHaveLength(5);
		await expect(within(pop).getByText('and 7 smaller')).toBeInTheDocument();
	}}
/>

<!-- One sentence covers both reasons an account is left out, because the
     member's question is the same: why the chart and the figures disagree. -->
<Story
	name="CoverageSentence"
	args={{
		coverage:
			'Not every account is in this chart. An account is left out when its bank shares balances only, or when its history is under 30 days.'
	}}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(
				'Not every account is in this chart. An account is left out when its bank shares balances only, or when its history is under 30 days.'
			)
		).toBeInTheDocument();
	}}
/>

<Story
	name="ShortHistory"
	args={{ points: [], shortHistory: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Balance · EUR')).toBeInTheDocument();
		await expect(
			canvas.getByText('A balance chart appears once there is a week of history.')
		).toBeInTheDocument();
		await expect(canvas.queryByRole('img')).not.toBeInTheDocument();
		await expect(canvas.queryByText('22 Jun to 20 Sep')).not.toBeInTheDocument();
	}}
/>

<!-- The day a transfer left. The line moved, so the row is here and says what
     kind of movement it was rather than being dropped. -->
<Story
	name="ATransferLeaving"
	args={{ marked: index('1 Sep') }}
	play={async ({ canvasElement }) => {
		const pop = within(popover(canvasElement)!);
		for (const words of [
			'1 Sep',
			'€10,912.40',
			'€500.00 less than 31 Aug',
			'Transfer to savings',
			'−€500.00',
			'Between your accounts'
		]) {
			await expect(pop.getByText(words)).toBeInTheDocument();
		}
		await expect(pop.queryByText('Unusual')).not.toBeInTheDocument();
		await expect(live(canvasElement)).toHaveTextContent(
			'1 Sep. €10,912.40. €500.00 less than 31 Aug. Transfer to savings −€500.00, Between your accounts.'
		);
	}}
/>
