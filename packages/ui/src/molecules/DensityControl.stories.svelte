<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import DensityControl, { DENSITY_KEY } from './DensityControl.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/DensityControl',
		component: DensityControl,
		tags: ['autodocs'],
		args: { onchange: fn() }
	});
</script>

<!-- Fine pointer: a mouse or trackpad. The control shows. -->
<Story
	name="Fine pointer"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('radiogroup')).toBeVisible();
	}}
/>

<!-- Choosing compact sets the attribute the stylesheet keys on, and it is
	 remembered — the choice has to survive between visits (Yt1rq). -->
<Story
	name="Choosing compact"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('radio', { name: 'Compact' }));

		await expect(args.onchange).toHaveBeenCalledWith('compact');
		await expect(document.documentElement).toHaveAttribute('data-density', 'compact');
		await expect(localStorage.getItem(DENSITY_KEY)).toBe('compact');
	}}
/>

<Story
	name="Choosing comfortable"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('radio', { name: 'Compact' }));
		await userEvent.click(canvas.getByRole('radio', { name: 'Comfortable' }));

		await expect(args.onchange).toHaveBeenLastCalledWith('comfortable');
		await expect(document.documentElement).toHaveAttribute('data-density', 'comfortable');
	}}
/>

<!-- Coarse pointer: `(any-pointer: coarse)` is a hardware media feature, which
     Storybook's own controls cannot emulate — this state is real on a touch
     device, where the control's own `display: none` takes over. Verified by
     reading the rule rather than a play function here. -->
<Story name="Coarse pointer (see the rule in DensityControl.svelte)" tags={['!test']} />
