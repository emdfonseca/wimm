<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, within } from 'storybook/test';
	import BalanceChart from './BalanceChart.svelte';
	import { balanceChart, balancePoints, balanceSummary } from './BalanceChart.fixture';

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

		const readout = canvasElement.querySelector('[aria-live="polite"]')!;
		await expect(readout.textContent).toBe('');

		await userEvent.keyboard('{ArrowLeft}');
		await expect(readout).toHaveTextContent(/^19 Sep · €/);
		await userEvent.keyboard('{ArrowRight}');
		await expect(readout).toHaveTextContent('20 Sep · €11,693.55');
		await userEvent.keyboard('{Home}');
		await expect(readout).toHaveTextContent('22 Jun · €9,640.00');
		await userEvent.keyboard('{ArrowLeft}');
		await expect(readout).toHaveTextContent('22 Jun · €9,640.00');
		await userEvent.keyboard('{End}');
		await expect(readout).toHaveTextContent('20 Sep · €11,693.55');
	}}
/>

<Story
	name="ADayPointedAt"
	args={{ marked: balancePoints.findIndex((p) => p.date === '3 Aug') }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('3 Aug · €9,870.12')).toBeInTheDocument();
		await expect(canvas.getByRole('img', { name: balanceSummary })).toBeInTheDocument();
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
