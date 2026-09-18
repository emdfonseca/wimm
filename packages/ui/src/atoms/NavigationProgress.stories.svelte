<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, userEvent, waitFor, within } from 'storybook/test';
	import NavigationProgress from './NavigationProgress.svelte';

	const { Story } = defineMeta({
		title: 'Atoms/NavigationProgress',
		component: NavigationProgress,
		tags: ['autodocs'],
		args: { active: false }
	});
</script>

<script lang="ts">
	// A play function has no `updateArgs`: that hook is for a story's render,
	// not for scripting one after the fact. Each of these toggles its own
	// local flag instead, through a button the play function clicks — real
	// prop reactivity, exercised the same way a member's browser would.
	let fastActive = $state(false);
	let waitingActive = $state(false);
	let minimumActive = $state(false);
</script>

<Story name="At rest" />

<!-- The 150ms delay: a navigation that starts and finishes before it elapses
     shows nothing at all, which is the flicker this avoids. -->
<Story
	name="A fast navigation shows nothing"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'start' }));
		await new Promise((r) => setTimeout(r, 60));
		await userEvent.click(canvas.getByRole('button', { name: 'stop' }));
		await new Promise((r) => setTimeout(r, 250));
		await expect(canvasElement.querySelector('.track')).toBeNull();
	}}
>
	{#snippet template()}
		<button type="button" onclick={() => (fastActive = true)}>start</button>
		<button type="button" onclick={() => (fastActive = false)}>stop</button>
		<NavigationProgress active={fastActive} />
	{/snippet}
</Story>

<!-- Waiting: past the 150ms delay, the bar appears and travels. -->
<Story
	name="Waiting"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'start' }));
		await waitFor(() => expect(canvasElement.querySelector('.track')).not.toBeNull(), {
			timeout: 1000
		});
	}}
>
	{#snippet template()}
		<button type="button" onclick={() => (waitingActive = true)}>start</button>
		<NavigationProgress active={waitingActive} />
	{/snippet}
</Story>

<!-- The 400ms minimum: a navigation that finishes just after the bar appeared
     keeps it on screen for the rest of that window rather than flashing it
     away immediately. -->
<Story
	name="The minimum-visible window holds"
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.click(canvas.getByRole('button', { name: 'start' }));
		await waitFor(() => expect(canvasElement.querySelector('.track')).not.toBeNull(), {
			timeout: 1000
		});
		await userEvent.click(canvas.getByRole('button', { name: 'stop' }));
		// Immediately after finishing, the bar is still required to be visible.
		await expect(canvasElement.querySelector('.track')).not.toBeNull();
		await waitFor(() => expect(canvasElement.querySelector('.track')).toBeNull(), {
			timeout: 1000
		});
	}}
>
	{#snippet template()}
		<button type="button" onclick={() => (minimumActive = true)}>start</button>
		<button type="button" onclick={() => (minimumActive = false)}>stop</button>
		<NavigationProgress active={minimumActive} />
	{/snippet}
</Story>

<!-- Reduced motion: the component's own `@media (prefers-reduced-motion:
     reduce)` rule is what ships this; forced here with a class only so the
     state is visible without an OS-level toggle, since Storybook has no such
     global configured. -->
<Story name="Reduced motion" tags={['!test']}>
	{#snippet template()}
		<div class="reduced-motion-demo">
			<NavigationProgress active={true} />
		</div>
	{/snippet}
</Story>

<style>
	:global(.reduced-motion-demo .segment) {
		animation: none !important;
		inline-size: 100% !important;
	}
</style>
