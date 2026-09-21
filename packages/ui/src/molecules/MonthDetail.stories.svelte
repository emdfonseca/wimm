<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, within } from 'storybook/test';
	import MonthDetail from './MonthDetail.svelte';

	const august = {
		id: 'detail',
		month: 'August 2026',
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€2,144.60' },
			{ label: 'Net', value: '+€305.40' }
		],
		risers: [
			{ label: 'Galp', value: '€246.80 · €82.40 more than usual', proportion: 1 },
			{ label: 'Amazon', value: '€188.20 · €64.10 more than usual', proportion: 0.78 },
			{ label: 'Zara', value: '€120.00 · not usually paid', proportion: 0.5 }
		]
	};

	const march = {
		month: 'March 2026',
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€3,930.30' },
			{ label: 'Net', value: '−€1,480.30' },
			{ label: 'Without unusual payments', value: '+€379.70' }
		],
		risers: [
			{ label: 'Auto Reparadora', value: '€1,650.00 · not usually paid', proportion: 1 },
			{ label: 'Galp', value: '€318.40 · €154.00 more than usual', proportion: 0.09 }
		],
		payments: [
			{
				description: 'Auto Reparadora',
				account: 'Current account',
				amount: '−€1,650.00',
				date: '12 Mar',
				negative: true,
				unusual: true,
				note: 'first payment to Auto Reparadora',
				href: '/transactions?page=2026-03-12.a1'
			},
			{
				description: 'Galp',
				account: 'Current account',
				amount: '−€210.00',
				date: '27 Mar',
				negative: true,
				unusual: true,
				note: 'usually about €60',
				href: '/transactions?page=2026-03-27.b2'
			}
		]
	};

	const may = {
		month: 'May 2026',
		risers: [],
		figures: [
			{ label: 'Money in', value: '€12,254.00' },
			{ label: 'Money out', value: '€2,225.00' },
			{ label: 'Net', value: '+€10,029.00' },
			{ label: 'Without unusual payments', value: '+€225.00' }
		],
		payments: [
			{
				description: 'Employer Lda',
				account: 'Current account',
				amount: '+€9,804.00',
				date: '22 May',
				negative: false,
				unusual: true,
				note: 'usually about €2,450',
				href: '/transactions?page=2026-05-22.c3'
			}
		]
	};

	const { Story } = defineMeta({
		title: 'Molecules/MonthDetail',
		component: MonthDetail,
		tags: ['autodocs'],
		parameters: { layout: 'padded' },
		args: august
	});
</script>

<Story
	name="Detail"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('region', { name: 'August 2026' })).toBeVisible();
		for (const words of [
			'Money in',
			'€2,450.00',
			'Money out',
			'€2,144.60',
			'Net',
			'More than usual',
			'Galp',
			'€246.80 · €82.40 more than usual',
			'Amazon',
			'€188.20 · €64.10 more than usual',
			'Zara',
			'€120.00 · not usually paid'
		]) {
			await expect(canvas.getAllByText(words)[0]).toBeVisible();
		}
		await expect(canvas.queryByText('Unusual payments')).not.toBeInTheDocument();
		await expect(canvas.queryByText(/unusual/)).not.toBeInTheDocument();
		await expect(canvas.queryByRole('button')).not.toBeInTheDocument();
	}}
/>

<Story
	name="DetailWithUnusual"
	args={march}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const words of [
			'Money in',
			'€3,930.30',
			'Without unusual payments',
			'+€379.70',
			'More than usual',
			'€1,650.00 · not usually paid',
			'€318.40 · €154.00 more than usual',
			'Unusual payments',
			'12 Mar',
			'Auto Reparadora',
			'−€1,650.00',
			'first payment to Auto Reparadora',
			'27 Mar',
			'−€210.00',
			'usually about €60'
		]) {
			await expect(canvas.getAllByText(words)[0]).toBeVisible();
		}
		await expect(canvas.getAllByText('Unusual')).toHaveLength(2);
		await expect(canvas.getAllByRole('link')).toHaveLength(2);
		await expect(canvas.queryByText(/usually about.*Auto/)).not.toBeInTheDocument();
	}}
/>

<Story
	name="DetailWithUnusualIncome"
	args={may}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		for (const words of [
			'€12,254.00',
			'€2,225.00',
			'Without unusual payments',
			'Unusual payments',
			'22 May',
			'Employer Lda',
			'+€9,804.00',
			'Unusual income',
			'usually about €2,450'
		]) {
			await expect(canvas.getAllByText(words)[0]).toBeVisible();
		}
		await expect(canvas.queryByText('More than usual')).not.toBeInTheDocument();
	}}
/>

<!-- Under way, so it is not set against whole months. -->
<Story
	name="DetailSoFar"
	args={{
		month: 'September 2026',
		span: '1 to 20 Sep',
		figures: [
			{ label: 'Money in', value: '€2,450.00' },
			{ label: 'Money out', value: '€1,812.64' },
			{ label: 'Net', value: '+€637.36' }
		],
		note: 'A month under way is not set against whole months.',
		risers: []
	}}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('region', { name: 'September 2026' })).toBeVisible();
		await expect(canvas.getByText('1 to 20 Sep')).toBeVisible();
		await expect(canvas.getByText('€1,812.64')).toBeVisible();
		await expect(
			canvas.getByText('A month under way is not set against whole months.')
		).toBeVisible();
		await expect(canvas.queryByText('More than usual')).not.toBeInTheDocument();
	}}
/>
