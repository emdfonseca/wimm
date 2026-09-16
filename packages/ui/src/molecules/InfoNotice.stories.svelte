<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import InfoNotice from './InfoNotice.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/InfoNotice',
		component: InfoNotice,
		tags: ['autodocs'],
		args: { title: 'Nothing was saved' }
	});
</script>

<!-- Polite by default. The member is returning from a native prompt, and
     interrupting them loses their place. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const notice = within(canvasElement).getByRole('status');
		await expect(notice).toHaveAttribute('aria-live', 'polite');
	}}
>
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<InfoNotice {...args}
				>You closed the prompt, so nothing changed. You can try again.</InfoNotice
			>
		</div>
	{/snippet}
</Story>

<Story name="You were signed out" args={{ title: 'You were signed out' }}>
	{#snippet template(args)}
		<div style="inline-size: 420px;">
			<InfoNotice {...args}>
				Your session ended. Sign in again and you will go back to what you were looking at.
			</InfoNotice>
		</div>
	{/snippet}
</Story>
