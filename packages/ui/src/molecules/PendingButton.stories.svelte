<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import PendingButton from './PendingButton.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/PendingButton',
		component: PendingButton,
		tags: ['autodocs'],
		argTypes: { pending: { control: 'boolean' } },
		args: {
			label: 'Create a passkey',
			pendingLabel: 'Waiting for your device',
			pending: false,
			onclick: fn()
		}
	});
</script>

<Story
	name="Rest"
	play={async ({ canvasElement, args }) => {
		const button = within(canvasElement).getByRole('button', { name: 'Create a passkey' });
		await expect(button).not.toHaveAttribute('aria-disabled');
		await userEvent.click(button);
		await expect(args.onclick).toHaveBeenCalledOnce();
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;"><PendingButton {...args} /></div>
	{/snippet}
</Story>

<!-- The whole point of the preset: while the browser has taken over the control
     carries its own label and does not fire. `aria-disabled` rather than
     `disabled`, so a member who tabbed to it keeps their place. -->
<Story
	name="Pending"
	args={{ pending: true }}
	play={async ({ canvasElement, args }) => {
		const button = within(canvasElement).getByRole('button', {
			name: 'Waiting for your device'
		});
		await expect(button).toHaveAttribute('aria-disabled', 'true');
		await expect(button).toHaveAttribute('aria-busy', 'true');
		await userEvent.click(button, { pointerEventsCheck: 0 });
		await expect(args.onclick).not.toHaveBeenCalled();
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;"><PendingButton {...args} /></div>
	{/snippet}
</Story>

<!-- A page that loads at rest announces nothing. The live region only speaks
     once the state has actually changed. -->
<Story
	name="Announces nothing at rest"
	play={async ({ canvasElement }) => {
		const status = within(canvasElement).getByRole('status');
		await expect(status).toHaveAttribute('aria-live', 'polite');
		await expect(status).toHaveTextContent('');
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;"><PendingButton {...args} /></div>
	{/snippet}
</Story>
