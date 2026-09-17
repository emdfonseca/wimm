<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import SegmentedControl from './SegmentedControl.svelte';

	const levels = [
		{ value: 'hidden', label: 'Nothing' },
		{ value: 'balance', label: 'Balance' },
		{ value: 'details', label: 'Details' }
	];

	const { Story } = defineMeta({
		title: 'Atoms/SegmentedControl',
		component: SegmentedControl,
		tags: ['autodocs'],
		argTypes: { disabled: { control: 'boolean' } },
		args: { options: levels, value: 'hidden', label: 'What Grace sees', onchange: fn() }
	});
</script>

<Story name="Nothing selected" />

<Story name="Balance selected" args={{ value: 'balance' }} />

<Story name="Details selected" args={{ value: 'details' }} />

<!-- One tab stop for the whole control: only the selected option is reachable
     by Tab, and the arrows move between them. -->
<Story
	name="One tab stop"
	args={{ value: 'balance' }}
	play={async ({ canvasElement }) => {
		const options = within(canvasElement).getAllByRole('radio');
		await expect(options[1]).toHaveAttribute('tabindex', '0');
		await expect(options[0]).toHaveAttribute('tabindex', '-1');
		await expect(options[2]).toHaveAttribute('tabindex', '-1');
	}}
/>

<!-- Arrow keys move the selection, not just focus: a control showing one thing
     and about to do another is the defect this avoids. -->
<Story
	name="Arrows move the selection"
	args={{ value: 'hidden' }}
	play={async ({ canvasElement, args }) => {
		const options = within(canvasElement).getAllByRole('radio');
		options[0]?.focus();

		await userEvent.keyboard('{ArrowRight}');
		await expect(args.onchange).toHaveBeenLastCalledWith('balance');

		await userEvent.keyboard('{ArrowRight}');
		await expect(args.onchange).toHaveBeenLastCalledWith('details');

		// And it wraps, so the control is never a dead end.
		await userEvent.keyboard('{ArrowRight}');
		await expect(args.onchange).toHaveBeenLastCalledWith('hidden');
	}}
/>

<Story
	name="Home and End"
	args={{ value: 'balance' }}
	play={async ({ canvasElement, args }) => {
		within(canvasElement).getAllByRole('radio')[1]?.focus();

		await userEvent.keyboard('{End}');
		await expect(args.onchange).toHaveBeenLastCalledWith('details');

		await userEvent.keyboard('{Home}');
		await expect(args.onchange).toHaveBeenLastCalledWith('hidden');
	}}
/>

<!-- Exactly one selection at all times: the property three checkboxes could
     not give. -->
<Story
	name="Exactly one is checked"
	args={{ value: 'details' }}
	play={async ({ canvasElement }) => {
		const checked = within(canvasElement)
			.getAllByRole('radio')
			.filter((option) => option.getAttribute('aria-checked') === 'true');
		await expect(checked).toHaveLength(1);
	}}
/>

<Story
	name="Clicking selects"
	play={async ({ canvasElement, args }) => {
		await userEvent.click(within(canvasElement).getByRole('radio', { name: 'Details' }));
		await expect(args.onchange).toHaveBeenCalledWith('details');
	}}
/>

<Story
	name="Disabled"
	args={{ value: 'balance', disabled: true }}
	play={async ({ canvasElement, args }) => {
		const group = within(canvasElement).getByRole('radiogroup');
		await expect(group).toHaveAttribute('aria-disabled', 'true');

		await userEvent.click(within(canvasElement).getByRole('radio', { name: 'Details' }), {
			pointerEventsCheck: 0
		});
		await expect(args.onchange).not.toHaveBeenCalled();
	}}
/>

<!-- The group names itself: three bare words are not self-describing. -->
<Story
	name="The group is named"
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('radiogroup')).toHaveAccessibleName(
			'What Grace sees'
		);
	}}
/>
