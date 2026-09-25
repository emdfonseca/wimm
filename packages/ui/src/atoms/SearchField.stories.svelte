<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, waitFor, within } from 'storybook/test';
	import SearchField from './SearchField.svelte';

	const placeholder = "Who it was with, or the bank's line";

	const { Story } = defineMeta({
		title: 'Atoms/SearchField',
		component: SearchField,
		tags: ['autodocs'],
		args: { label: 'Search', placeholder, name: 'q', onsearch: fn() }
	});
</script>

<!-- Named by its label and bounded at 100. Enter applies the search at once,
     trimmed, and the pause that would have applied it does nothing more. -->
<Story
	name="Empty"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		const box = canvas.getByRole('searchbox', { name: 'Search' });
		await expect(box).toHaveAttribute('placeholder', placeholder);
		await expect(box).toHaveAttribute('maxlength', '100');
		await expect(canvas.queryByRole('button', { name: 'Clear search' })).toBeNull();

		await userEvent.type(box, '  galp ');
		await expect(args.onsearch).not.toHaveBeenCalled();

		await userEvent.keyboard('{Enter}');
		await expect(args.onsearch).toHaveBeenCalledTimes(1);
		await expect(args.onsearch).toHaveBeenCalledWith('galp', { live: false });

		await new Promise((done) => setTimeout(done, 500));
		await expect(args.onsearch).toHaveBeenCalledTimes(1);
	}}
/>

<!-- The search applies as the member types, once they pause: one report for
     the word, not one per letter, and marked as coming from typing. -->
<Story
	name="Typing"
	play={async ({ canvasElement, args }) => {
		const box = within(canvasElement).getByRole('searchbox', { name: 'Search' });
		await userEvent.type(box, 'galp ');
		await waitFor(() => expect(args.onsearch).toHaveBeenCalledTimes(1));
		await expect(args.onsearch).toHaveBeenCalledWith('galp', { live: true });

		// A trailing space changes nothing searched, so nothing is reported.
		await userEvent.type(box, ' ');
		await new Promise((done) => setTimeout(done, 500));
		await expect(args.onsearch).toHaveBeenCalledTimes(1);
		await expect(box).toHaveValue('galp  ');
	}}
/>

<!-- Clearing is one action: the box empties, the empty search is applied, and
     focus goes back to the box rather than being lost with the button. -->
<Story
	name="WithText"
	args={{ value: 'galp' }}
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		const box = canvas.getByRole('searchbox', { name: 'Search' });
		await expect(box).toHaveValue('galp');

		await userEvent.click(canvas.getByRole('button', { name: 'Clear search' }));
		await expect(args.onsearch).toHaveBeenCalledTimes(1);
		await expect(args.onsearch).toHaveBeenCalledWith('', { live: false });
		await expect(box).toHaveValue('');
		await expect(box).toHaveFocus();
	}}
/>

<Story
	name="LongText"
	args={{ value: 'COMPRA 4829 GALP ENERGIA AV DA REPUBLICA LISBOA 230002268264350 PT' }}
	play={async ({ canvasElement }) => {
		const box = within(canvasElement).getByRole('searchbox', { name: 'Search' });
		await expect(box).toHaveValue(
			'COMPRA 4829 GALP ENERGIA AV DA REPUBLICA LISBOA 230002268264350 PT'
		);
	}}
/>
