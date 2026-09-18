<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, within } from 'storybook/test';
	import StepActions from './StepActions.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/StepActions',
		component: StepActions,
		tags: ['autodocs'],
		args: {
			primaryLabel: 'Continue to Monzo',
			lesserLabel: 'Cancel',
			onPrimary: fn(),
			onLesser: fn()
		}
	});
</script>

<!-- Medium and above: a row, lesser first, primary last. -->
<Story name="Row" parameters={{ viewport: { defaultViewport: 'wide' } }} />

<!-- Compact: stacked, primary on top. -->
<Story name="Stacked" parameters={{ viewport: { defaultViewport: 'compact' } }} />

<Story name="Pending" args={{ primaryPending: true }} />

<!-- Absent, not disabled: a step whose way forward is a choice made
     elsewhere on the screen carries only the lesser action. -->
<Story
	name="Lesser only"
	args={{ primaryLabel: undefined }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByRole('button', { name: /Continue/ })).toBeNull();
		await expect(canvas.getByRole('button', { name: 'Cancel' })).toBeInTheDocument();
	}}
/>
