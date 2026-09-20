<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import EnrolScreen from './EnrolScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/EnrolScreen',
		component: EnrolScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		argTypes: {
			state: {
				control: 'inline-radio',
				options: ['default', 'creating', 'not-saved', 'dismissed']
			}
		},
		args: { firstName: 'Ada', lastName: 'Lovelace', state: 'default', onenrol: fn() }
	});
</script>

<!-- J01.A / 01 · Default. The member is shown the name they were registered
     under before being asked to create anything. -->
<Story
	name="Default"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { level: 1 })).toHaveFocus();
		await expect(canvas.getByRole('heading', { level: 1 })).toHaveTextContent(
			'Hello, Ada Lovelace'
		);
		await userEvent.click(canvas.getByRole('button', { name: 'Create a passkey' }));
		await expect(args.onenrol).toHaveBeenCalledOnce();
	}}
/>

<!-- J01.A / 02 · Waiting. The browser has taken over; the control is inert and
     says so. -->
<Story
	name="Creating"
	args={{ state: 'creating' }}
	play={async ({ canvasElement, args }) => {
		const button = within(canvasElement).getByRole('button', {
			name: 'Waiting for your device…'
		});
		await expect(button).toHaveAttribute('aria-disabled', 'true');
		await userEvent.click(button, { pointerEventsCheck: 0 });
		await expect(args.onenrol).not.toHaveBeenCalled();
	}}
/>

<!-- The error is assertive and does not move focus: the member is returning
     from a native prompt. -->
<Story
	name="Passkey not saved"
	args={{ state: 'not-saved' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const alert = canvas.getByRole('alert');
		await expect(alert).toHaveAttribute('aria-live', 'assertive');
		await expect(alert).not.toHaveFocus();
		await expect(alert).toHaveTextContent('Your device did not save the passkey');
		await expect(alert).toHaveTextContent(
			/wimm needs a passkey your device can offer you next time/
		);
		// A refusal changes the action's own sentence, and the page says the link
		// still works.
		await expect(canvas.getByRole('button', { name: 'Try again' })).toBeEnabled();
		await expect(canvas.getByText('Your link still works.')).toBeInTheDocument();
	}}
/>

<!-- Closing the prompt changed nothing, so it waits rather than interrupts. -->
<Story
	name="Prompt dismissed"
	args={{ state: 'dismissed' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		// The button keeps an empty live region of its own so a later message has
		// somewhere to land, so the notice is found by what it says rather than by
		// being the only status on the screen.
		const status = canvas.getByText('Nothing was created').closest('[role="status"]');
		await expect(status).toHaveAttribute('aria-live', 'polite');
		await expect(canvas.getByRole('button', { name: 'Create a passkey' })).toBeEnabled();
	}}
/>
