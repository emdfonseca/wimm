<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import Notice from './Notice.svelte';

	const { Story } = defineMeta({
		title: 'Atoms/Notice',
		component: Notice,
		tags: ['autodocs'],
		argTypes: {
			title: { control: 'text' },
			live: { control: 'inline-radio', options: ['assertive', 'polite', 'off'] }
		},
		args: { title: 'Groceries is over budget', live: 'polite' }
	});
</script>

<!-- Notice carries no tone of its own: the tonal presets fix the colours, so a
     feedback colour is chosen in one place. Bare, it falls back to warning. -->
<Story name="Default">
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<Notice {...args}>$412.60 of $350.00 spent, with 9 days left in the period.</Notice>
		</div>
	{/snippet}
</Story>

<Story name="Title only" args={{ title: 'Your link still works' }}>
	{#snippet template(args)}
		<div style="inline-size: 420px;"><Notice {...args} /></div>
	{/snippet}
</Story>

<!-- Polite waits for a pause; assertive interrupts. Both are announced without
     moving focus. -->
<Story
	name="Polite is a status"
	play={async ({ canvasElement }) => {
		const notice = within(canvasElement).getByRole('status');
		await expect(notice).toHaveAttribute('aria-live', 'polite');
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<Notice {...args}>Nothing changed.</Notice>
		</div>
	{/snippet}
</Story>

<Story
	name="Assertive is an alert"
	args={{ live: 'assertive' }}
	play={async ({ canvasElement }) => {
		const notice = within(canvasElement).getByRole('alert');
		await expect(notice).toHaveAttribute('aria-live', 'assertive');
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<Notice {...args}>Your device did not save a passkey.</Notice>
		</div>
	{/snippet}
</Story>
