<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import AuthShell from './AuthShell.svelte';
	import Button from '../atoms/Button.svelte';
	import ErrorNotice from '../molecules/ErrorNotice.svelte';

	const { Story } = defineMeta({
		title: 'Templates/AuthShell',
		component: AuthShell,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { heading: 'Hello, Ana Reis' }
	});
</script>

<!-- Focus moves to the heading on load, not to the action: the member has to
     read who the page thinks they are before committing to anything. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const heading = within(canvasElement).getByRole('heading', { level: 1 });
		await expect(heading).toHaveFocus();
		await expect(heading).toHaveAttribute('tabindex', '-1');
	}}
>
	{#snippet template(args)}
		<AuthShell {...args}>
			{#snippet body()}
				A wimm account has been set up for you. Create a passkey and you are in.
			{/snippet}
			{#snippet action()}
				<Button block size="lg">Create a passkey</Button>
			{/snippet}
		</AuthShell>
	{/snippet}
</Story>

<!-- A result sits directly above the action that produced it, and does not
     take focus away from the heading. -->
<Story
	name="With a result"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('alert')).toBeInTheDocument();
		await expect(canvas.getByRole('heading', { level: 1 })).toHaveFocus();
	}}
>
	{#snippet template(args)}
		<AuthShell {...args}>
			{#snippet result()}
				<ErrorNotice title="Your device did not save the passkey">
					wimm needs a passkey your device can offer you next time.
				</ErrorNotice>
			{/snippet}
			{#snippet body()}
				A wimm account has been set up for you. Create a passkey and you are in.
			{/snippet}
			{#snippet action()}
				<Button block size="lg">Create a passkey</Button>
			{/snippet}
		</AuthShell>
	{/snippet}
</Story>

<!-- Every slot is settable by prop, and each is optional. -->
<Story name="Every slot" tags={['!test']}>
	{#snippet template(args)}
		<AuthShell {...args}>
			{#snippet result()}
				<ErrorNotice title="Result slot">Directly above the action.</ErrorNotice>
			{/snippet}
			{#snippet body()}Body slot.{/snippet}
			{#snippet action()}
				<Button block size="lg">Action slot</Button>
			{/snippet}
			{#snippet footnote()}Footnote slot.{/snippet}
		</AuthShell>
	{/snippet}
</Story>

<!-- At Compact the 560 panel becomes a 200 band above a 342 card. Select the
     Compact viewport to see it; the structure is the only difference. -->
<Story name="Heading only" tags={['!test']}>
	{#snippet template(args)}
		<AuthShell {...args} />
	{/snippet}
</Story>
