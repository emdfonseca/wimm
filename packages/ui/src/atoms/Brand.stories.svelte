<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Brand from './Brand.svelte';

	const { Story } = defineMeta({
		title: 'Atoms/Brand',
		component: Brand,
		tags: ['autodocs'],
		argTypes: {
			inverted: { control: 'boolean' },
			markOnly: { control: 'boolean' }
		},
		args: { inverted: false, markOnly: false }
	});
</script>

<Story name="Default" />

<Story name="Mark only" args={{ markOnly: true }} />

<!-- On the brand panel the stock mark would be brand green on a brand-green
     ground, so the tile takes the light colour and the glyph takes the brand. -->
<Story name="Inverted" args={{ inverted: true }}>
	{#snippet template(args)}
		<div
			style="padding: 24px; background: linear-gradient(160deg, var(--gradient-brand-from), var(--gradient-brand-to));"
		>
			<Brand {...args} />
		</div>
	{/snippet}
</Story>
