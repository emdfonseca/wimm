<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import Button from './Button.svelte';

	const { Story } = defineMeta({
		title: 'Atoms/Button',
		component: Button,
		tags: ['autodocs'],
		argTypes: {
			variant: {
				control: 'select',
				options: ['primary', 'secondary', 'ghost', 'destructive']
			},
			size: { control: 'inline-radio', options: ['sm', 'md', 'lg'] },
			block: { control: 'boolean' },
			disabled: { control: 'boolean' }
		},
		args: { variant: 'primary', size: 'md', onclick: fn() }
	});
</script>

<Story name="Primary">
	{#snippet template({ children: _label, ...args })}
		<Button {...args}>Create a passkey</Button>
	{/snippet}
</Story>

<Story name="Secondary" args={{ variant: 'secondary' }}>
	{#snippet template({ children: _label, ...args })}
		<Button {...args}>Cancel</Button>
	{/snippet}
</Story>

<Story name="Ghost" args={{ variant: 'ghost' }}>
	{#snippet template({ children: _label, ...args })}
		<Button {...args}>Dismiss</Button>
	{/snippet}
</Story>

<Story name="Destructive" args={{ variant: 'destructive' }}>
	{#snippet template({ children: _label, ...args })}
		<Button {...args}>Remove passkey</Button>
	{/snippet}
</Story>

<!-- Disabled is the leftmost state in the precedence chain: it owns the fill
     and the border whatever the variant was, and it does not fire. -->
<Story
	name="Disabled"
	args={{ disabled: true }}
	play={async ({ canvasElement, args }) => {
		const button = within(canvasElement).getByRole('button');
		await expect(button).toBeDisabled();
		await userEvent.click(button, { pointerEventsCheck: 0 });
		await expect(args.onclick).not.toHaveBeenCalled();
	}}
>
	{#snippet template({ children: _label, ...args })}
		<Button {...args}>Create a passkey</Button>
	{/snippet}
</Story>

<!-- Every size clears the 24 px target floor of SC 2.5.8. -->
<Story name="Sizes" tags={['!test']}>
	{#snippet template({ children: _label, ...args })}
		<div style="display: flex; gap: 12px; align-items: center;">
			<Button {...args} size="sm">Small</Button>
			<Button {...args} size="md">Medium</Button>
			<Button {...args} size="lg">Large</Button>
		</div>
	{/snippet}
</Story>

<Story
	name="Fires on click"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('button'));
		await expect(args.onclick).toHaveBeenCalledOnce();
	}}
>
	{#snippet template({ children: _label, ...args })}
		<Button {...args}>Create a passkey</Button>
	{/snippet}
</Story>
