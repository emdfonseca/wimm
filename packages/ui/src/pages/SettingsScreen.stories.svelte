<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import SettingsScreen from './SettingsScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/SettingsScreen',
		component: SettingsScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/settings' },
		args: { ontheme: fn(), ondensity: fn() }
	});
</script>

<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { name: 'Settings', level: 1 })).toBeInTheDocument();
		await expect(canvas.getByRole('heading', { name: 'Appearance', level: 2 })).toBeInTheDocument();
		await expect(canvas.getByRole('radiogroup', { name: 'Theme' })).toBeInTheDocument();
		await expect(canvas.getByRole('radiogroup', { name: 'Rows' })).toBeInTheDocument();
	}}
/>

<Story
	name="Choosing a theme"
	play={async ({ canvasElement, args }) => {
		const group = within(canvasElement).getByRole('radiogroup', { name: 'Theme' });
		await userEvent.click(within(group).getByRole('radio', { name: 'Dark' }));
		await expect(args.ontheme).toHaveBeenCalledWith('dark');
	}}
/>

<Story
	name="Choosing row height"
	play={async ({ canvasElement, args }) => {
		const group = within(canvasElement).getByRole('radiogroup', { name: 'Rows' });
		await userEvent.click(within(group).getByRole('radio', { name: 'Compact' }));
		await expect(args.ondensity).toHaveBeenCalledWith('compact');
	}}
/>

<!-- ADR 0004: on a touch screen the row-height control is hidden entirely,
     never disabled, and the helper says why rather than leaving a gap.
     `(any-pointer: coarse)` cannot be forced from Storybook's own controls, so
     which helper is showing cannot be asserted here. What can be held is that
     the words a touch screen shows are in the page for the media query to
     reveal. -->
<Story
	name="On a touch screen (see the rule in SettingsScreen.svelte)"
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(
				/Row height is not offered here\..*on a\s+touch screen wimm stays comfortable/
			)
		).toBeInTheDocument();
	}}
/>
