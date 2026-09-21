<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, within } from 'storybook/test';
	import MonthlyNetChart from './MonthlyNetChart.svelte';
	import {
		heldMonths,
		netChart,
		netMonths,
		netSummary,
		twoFullMonths,
		twoFullMonthsSummary
	} from './MonthlyNetChart.fixture';

	const live = (root: Element) => root.querySelector('[aria-live="polite"]')!;
	const rect = (root: Element, month: string) => root.querySelector(`rect[data-month="${month}"]`)!;

	const { Story } = defineMeta({
		title: 'Molecules/MonthlyNetChart',
		component: MonthlyNetChart,
		tags: ['autodocs'],
		args: netChart
	});
</script>

<!-- One tab stop the keyboard can walk, and a name that is a sentence. -->
<Story
	name="AsItOpens"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Month by month · EUR')).toBeInTheDocument();
		await expect(canvas.getByText('September 2025 to September 2026')).toBeInTheDocument();
		await expect(canvas.getByText('+€637')).toBeInTheDocument();
		await expect(canvas.getByText('−€1,480')).toBeInTheDocument();

		const plot = canvas.getByRole('img', { name: netSummary });
		await userEvent.tab();
		await expect(plot).toHaveFocus();
		await expect(live(canvasElement).textContent).toBe('');

		await userEvent.keyboard('{ArrowLeft}');
		await expect(live(canvasElement)).toHaveTextContent('August 2026 · +€305.40');
		await userEvent.keyboard('{ArrowRight}');
		await expect(live(canvasElement)).toHaveTextContent('September 2026 · +€637.36');
		await userEvent.keyboard('{ArrowRight}');
		await expect(live(canvasElement)).toHaveTextContent('September 2026 · +€637.36');
		await userEvent.keyboard('{Home}');
		await expect(live(canvasElement)).toHaveTextContent('September 2025 · +€212.40');
		await userEvent.keyboard('{ArrowLeft}');
		await expect(live(canvasElement)).toHaveTextContent('September 2025 · +€212.40');
		await userEvent.keyboard('{End}');
		await expect(live(canvasElement)).toHaveTextContent('September 2026 · +€637.36');
		await userEvent.keyboard(
			'{Home}{ArrowRight}{ArrowRight}{ArrowRight}{ArrowRight}{ArrowRight}{ArrowRight}'
		);
		await expect(live(canvasElement)).toHaveTextContent('March 2026 · −€1,480.30');
	}}
/>

<Story
	name="AMonthPointedAt"
	args={{ marked: netMonths.findIndex((m) => m.label === 'March 2026') }}
	play={async ({ canvasElement }) => {
		await expect(live(canvasElement)).toHaveTextContent('March 2026 · −€1,480.30');
		await expect(within(canvasElement).getByRole('img', { name: netSummary })).toBeInTheDocument();
	}}
/>

<!-- Direction is which side of the zero rule, and the sign says it too. -->
<Story
	name="AMonthBelowZero"
	args={{ marked: netMonths.findIndex((m) => m.label === 'March 2026') }}
	play={async ({ canvasElement }) => {
		const zero = canvasElement.querySelector('.zero')!.getBoundingClientRect();
		const below = rect(canvasElement, 'March 2026').getBoundingClientRect();
		const above = rect(canvasElement, 'February 2026').getBoundingClientRect();
		await expect(below.top).toBeGreaterThanOrEqual(zero.top - 1);
		await expect(above.bottom).toBeLessThanOrEqual(zero.bottom + 1);
		await expect(live(canvasElement)).toHaveTextContent('March 2026 · −€1,480.30');
	}}
/>

<!-- A month under way and a month a ledger begins inside are not set beside
     whole months: both are drawn lighter. -->
<Story
	name="PartlyHeldMonths"
	args={{ months: heldMonths }}
	play={async ({ canvasElement }) => {
		const partial = [...canvasElement.querySelectorAll('rect.partial')].map((r) =>
			r.getAttribute('data-month')
		);
		await expect(partial).toEqual(['April 2026', 'September 2026']);
		await expect(rect(canvasElement, 'May 2026').classList.contains('partial')).toBe(false);
	}}
/>

<Story
	name="NoTypicalMonthYet"
	args={{
		months: twoFullMonths,
		typical: undefined,
		high: '+€637',
		low: '−€60',
		summary: twoFullMonthsSummary
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('img', { name: twoFullMonthsSummary })).toBeInTheDocument();
		await expect(canvasElement.querySelector('.typical')).toBeNull();
		await expect(twoFullMonthsSummary).not.toMatch(/typical/);
	}}
/>
