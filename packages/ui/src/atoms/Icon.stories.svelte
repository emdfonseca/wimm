<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Icon from './Icon.svelte';

	const { Story } = defineMeta({
		title: 'Atoms/Icon',
		component: Icon,
		tags: ['autodocs'],
		argTypes: {
			name: {
				control: 'select',
				options: [
					'circle-help',
					'shield-check',
					'smartphone',
					'layout-dashboard',
					'user-round-check',
					'panel-left-close',
					'arrow-left-right',
					'settings'
				]
			},
			size: { control: { type: 'number' } }
		},
		args: { name: 'shield-check', size: 24 }
	});
</script>

<Story name="Default" />

<!-- Every glyph the canvas names, at the sizes it draws them: 16 in the
     sidebar, 20 on the brand panel, 24 in an empty state. -->
<Story name="All glyphs" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; gap: 20px; align-items: center; color: var(--color-text-secondary);">
			<Icon {...args} name="circle-help" />
			<Icon {...args} name="shield-check" />
			<Icon {...args} name="smartphone" />
			<Icon {...args} name="layout-dashboard" />
			<Icon {...args} name="user-round-check" />
			<Icon {...args} name="panel-left-close" />
			<Icon {...args} name="arrow-left-right" />
			<Icon {...args} name="settings" />
		</div>
	{/snippet}
</Story>

<Story name="Sizes" tags={['!test']}>
	{#snippet template(args)}
		<div style="display: flex; gap: 16px; align-items: center;">
			<Icon {...args} size={16} />
			<Icon {...args} size={20} />
			<Icon {...args} size={24} />
		</div>
	{/snippet}
</Story>
