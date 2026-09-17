<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import ConsentExplainerScreen from './ConsentExplainerScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/ConsentExplainerScreen',
		component: ConsentExplainerScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: {
			bankName: 'Montepio',
			accessEndsOn: '16 December 2026',
			oncontinue: fn(),
			oncancel: fn()
		}
	});
</script>

<!-- 90 days, which is what Montepio and Revolut grant. -->
<Story
	name="Default"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('16 December 2026')).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Continue to Montepio' }));
		await expect(args.oncontinue).toHaveBeenCalledOnce();
	}}
/>

<!-- ActivoBank grants a single day. A member not shown that date reads the
     daily prompt as a defect rather than as their bank's limit. -->
<Story
	name="A bank that grants one day"
	args={{ bankName: 'ActivoBank', accessEndsOn: 'tomorrow', shortLived: true }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('tomorrow')).toBeInTheDocument();
		await expect(canvas.getByText(/That is soon/)).toBeInTheDocument();
	}}
/>

<!-- The date is the bank's own limit, and the copy says so: otherwise a member
     reads it as something wimm chose and could change. -->
<Story
	name="The date belongs to the bank"
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/Montepio's own limit rather than a wimm setting/)
		).toBeInTheDocument();
	}}
/>

<!-- Transactions are out of scope in this change, and the consent asked for
     says so rather than quietly requesting them for later. -->
<Story
	name="It states what is not shared"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/Not your transactions/)).toBeInTheDocument();
		await expect(canvas.getByText(/cannot make a payment/)).toBeInTheDocument();
		await expect(canvas.getByText(/never sees it/)).toBeInTheDocument();
	}}
/>

<Story
	name="Cancelling"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Cancel' }));
		await expect(args.oncancel).toHaveBeenCalledOnce();
		await expect(args.oncontinue).not.toHaveBeenCalled();
	}}
/>
