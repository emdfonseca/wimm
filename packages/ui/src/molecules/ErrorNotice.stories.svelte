<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import ErrorNotice from './ErrorNotice.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/ErrorNotice',
		component: ErrorNotice,
		tags: ['autodocs'],
		args: { title: 'Your device did not save a passkey' }
	});
</script>

<!-- Assertive by default: this stops the member getting what they came for. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const notice = within(canvasElement).getByRole('alert');
		await expect(notice).toHaveAttribute('aria-live', 'assertive');
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<ErrorNotice {...args}>
				It needs to save one you can sign in with without typing anything. Your link still works.
			</ErrorNotice>
		</div>
	{/snippet}
</Story>

<Story name="Passkey not recognised" args={{ title: 'This wimm has no record of it' }}>
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<ErrorNotice {...args}>Nothing was signed in.</ErrorNotice>
		</div>
	{/snippet}
</Story>
