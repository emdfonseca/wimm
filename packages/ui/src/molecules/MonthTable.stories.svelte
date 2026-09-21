<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, within } from 'storybook/test';
	import MonthTable from './MonthTable.svelte';

	const rows = [
		{
			key: 'sep',
			label: 'September 2026',
			state: 'So far',
			moneyIn: '€2,450.00',
			moneyOut: '€1,812.64',
			net: '+€637.36'
		},
		{ key: 'aug', label: 'August 2026', moneyIn: '€2,450.00', moneyOut: '€2,144.60', net: '+€305.40' },
		{
			key: 'mar',
			label: 'March 2026',
			moneyIn: '€2,450.00',
			moneyOut: '€3,930.30',
			net: '−€1,480.30',
			unusualLine: '2 unusual · +€379.70 without them'
		},
		{
			key: 'apr',
			label: 'April 2026',
			state: 'Held from 12 Apr',
			moneyIn: '€120.00',
			moneyOut: '€78.90',
			net: '+€41.10'
		}
	];

	const { Story } = defineMeta({
		title: 'Molecules/MonthTable',
		component: MonthTable,
		tags: ['autodocs'],
		parameters: { layout: 'padded' },
		args: { label: 'Months', rows, selected: 'aug' }
	});
</script>

<!-- One choice at all times, made by Enter or Space, and focus stays on the
     row: the detail it names is somewhere else on the page. -->
<Story
	name="Row"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const button = canvas.getByRole('button', { name: /August 2026/ });
		await expect(button).toHaveTextContent('+€305.40');
		await expect(canvas.getByText('+€305.40', { selector: '.net' })).toBeInTheDocument();
		await expect(button).not.toHaveAttribute('aria-controls');
		await expect(button.getBoundingClientRect().height).toBe(56);
		await expect(canvas.getAllByRole('button')).toHaveLength(4);
	}}
/>

<Story
	name="RowWithUnusual"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const button = canvas.getByRole('button', { name: /March 2026/ });
		await expect(button).toHaveTextContent('−€1,480.30');
		await expect(button).toHaveTextContent('2 unusual · +€379.70 without them');
		await expect(canvas.getByRole('button', { name: /August 2026/ })).not.toHaveTextContent(
			/unusual/
		);
	}}
/>

<Story
	name="Selected"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('button', { name: /August 2026/ })).toHaveAttribute(
			'aria-pressed',
			'true'
		);
		for (const other of [/September 2026/, /March 2026/, /April 2026/]) {
			await expect(canvas.getByRole('button', { name: other })).toHaveAttribute(
				'aria-pressed',
				'false'
			);
		}
	}}
/>

<Story
	name="SoFar"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const button = canvas.getByRole('button', { name: /September 2026/ });
		await expect(button).toHaveTextContent('So far');
		await expect(button).toHaveTextContent('+€637.36');
	}}
/>

<Story
	name="HeldFrom"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const button = canvas.getByRole('button', { name: /April 2026/ });
		await expect(button).toHaveTextContent('Held from 12 Apr');
		await expect(button).toHaveTextContent('+€41.10');
	}}
/>

<Story
	name="Choosing"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const march = canvas.getByRole('button', { name: /March 2026/ });
		const august = canvas.getByRole('button', { name: /August 2026/ });

		await userEvent.tab();
		await userEvent.tab();
		await userEvent.tab();
		await expect(march).toHaveFocus();
		await userEvent.keyboard('{Enter}');
		await expect(march).toHaveAttribute('aria-pressed', 'true');
		await expect(august).toHaveAttribute('aria-pressed', 'false');
		await expect(march).toHaveFocus();

		await userEvent.tab({ shift: true });
		await expect(august).toHaveFocus();
		await userEvent.keyboard(' ');
		await expect(august).toHaveAttribute('aria-pressed', 'true');
		await expect(march).toHaveAttribute('aria-pressed', 'false');
		await expect(august).toHaveFocus();

		// Choosing the chosen one changes nothing.
		await userEvent.keyboard('{Enter}');
		await expect(august).toHaveAttribute('aria-pressed', 'true');
	}}
/>
