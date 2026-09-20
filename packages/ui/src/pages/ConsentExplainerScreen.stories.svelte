<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import ConsentExplainerScreen from './ConsentExplainerScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/ConsentExplainerScreen',
		component: ConsentExplainerScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/connect' },
		args: {
			bankName: 'Monzo',
			oncontinue: fn(),
			oncancel: fn()
		}
	});
</script>

<Story
	name="Default"
	tags={['kind-state']}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('wimm will read your Monzo accounts')).toBeInTheDocument();
		await expect(canvas.getByRole('link', { name: 'Pick another bank' })).toBeInTheDocument();
		await userEvent.click(canvas.getByRole('button', { name: 'Continue to Monzo' }));
		await expect(args.oncontinue).toHaveBeenCalledOnce();
	}}
/>

<!-- What is read, and the two things that are not. -->
<Story
	name="It states what is and is not read"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText(/last four digits of its number/)).toBeInTheDocument();
		await expect(canvas.getByText(/read now and again/)).toBeInTheDocument();
		await expect(canvas.getByText(/does not see your transactions/)).toBeInTheDocument();
		await expect(canvas.getByText(/never sees your Monzo password/)).toBeInTheDocument();
	}}
/>

<!-- Privacy is per person and reversible, stated before the member can no
     longer change what they granted here. -->
<Story
	name="It says who can see it"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Only you can see these accounts')).toBeInTheDocument();
		await expect(canvas.getByText(/nothing here reaches another member/)).toBeInTheDocument();
	}}
/>

<Story
	name="It says the member is about to leave"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText('You will leave wimm and come back here when Monzo is done.')
		).toBeInTheDocument();
	}}
/>

<Story
	name="Cancelling"
	tags={['kind-behaviour']}
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button', { name: 'Cancel' }));
		await expect(args.oncancel).toHaveBeenCalledOnce();
		await expect(args.oncontinue).not.toHaveBeenCalled();
	}}
/>

<!-- Compact says the same thing in one sentence: the notice is the only part
     of this screen whose words change with the regime. -->
<Story
	name="Compact"
	tags={['kind-state', 'size-compact']}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Only you can see these accounts')).toBeInTheDocument();
		await expect(
			canvas.getByText(/Nobody else sees an account until you choose who does/)
		).toBeInTheDocument();
	}}
/>
