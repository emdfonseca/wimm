<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import ThemeToggle, { THEME_KEY } from './ThemeToggle.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/ThemeToggle',
		component: ThemeToggle,
		tags: ['autodocs'],
		args: { onchange: fn() }
	});
</script>

<!-- Three states, because the stylesheet has three: light and dark resolved
     from prefers-color-scheme by default, and either forced with [data-theme].
     A two-way switch would make "follow the system" unreachable. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('radiogroup')).toHaveAccessibleName('Appearance');
		await expect(canvas.getAllByRole('radio')).toHaveLength(3);
	}}
/>

<!-- Choosing dark sets the attribute the stylesheet keys on. -->
<Story
	name="Choosing dark"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('radio', { name: 'Dark' }));

		await expect(args.onchange).toHaveBeenCalledWith('dark');
		await expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
		await expect(localStorage.getItem(THEME_KEY)).toBe('dark');
	}}
/>

<Story
	name="Choosing light"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('radio', { name: 'Light' }));

		await expect(args.onchange).toHaveBeenCalledWith('light');
		await expect(document.documentElement).toHaveAttribute('data-theme', 'light');
	}}
/>

<!-- System removes the attribute rather than writing a value: its absence is
     what lets the media query decide, so writing "system" would pin the theme
     to whatever it was. -->
<Story
	name="Back to system"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);

		await userEvent.click(canvas.getByRole('radio', { name: 'Dark' }));
		await expect(document.documentElement).toHaveAttribute('data-theme', 'dark');

		await userEvent.click(canvas.getByRole('radio', { name: 'System' }));
		await expect(args.onchange).toHaveBeenLastCalledWith('system');
		await expect(document.documentElement).not.toHaveAttribute('data-theme');
		await expect(localStorage.getItem(THEME_KEY)).toBe('system');
	}}
/>

<!-- One control, one tab stop, arrows between the options — inherited from the
     segmented control it is built on. -->
<Story
	name="Keyboard"
	play={async ({ canvasElement, args }) => {
		const options = within(canvasElement).getAllByRole('radio');
		options[0]?.focus();
		await userEvent.keyboard('{ArrowRight}');

		await expect(args.onchange).toHaveBeenCalledWith('light');
	}}
/>
