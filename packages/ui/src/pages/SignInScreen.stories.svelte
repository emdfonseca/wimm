<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import SignInScreen from './SignInScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/SignInScreen',
		component: SignInScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		argTypes: {
			state: {
				control: 'inline-radio',
				options: ['default', 'signing-in', 'dismissed', 'expired', 'not-recognised']
			}
		},
		args: { state: 'default', onsignin: fn() }
	});
</script>

<!-- One control and no identifier field: the credential is discoverable, so
     there is nothing to type and nothing to look up. -->
<Story
	name="Default"
	play={async ({ canvasElement, args }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { level: 1 })).toHaveFocus();
		await expect(canvas.queryByRole('textbox')).toBeNull();
		await userEvent.click(canvas.getByRole('button', { name: 'Sign in with a passkey' }));
		await expect(args.onsignin).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="Signing in"
	args={{ state: 'signing-in' }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByRole('button', { name: 'Waiting for your device…' })
		).toHaveAttribute('aria-disabled', 'true');
	}}
/>

<Story
	name="Prompt dismissed"
	args={{ state: 'dismissed' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByText('Nothing happened')).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Sign in with a passkey' })).toBeEnabled();
	}}
/>

<!-- An expired session sends the member here rather than to a dead end, and the
     page promises they will go back to what they were looking at. -->
<Story
	name="Session expired"
	args={{ state: 'expired' }}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByText('You were signed out')).toBeInTheDocument();
	}}
/>

<!-- J02.B. The member offered a credential this instance has no record of. It
     is not a page of its own: the control stays live, because offering the
     wrong passkey is recovered by offering another one. -->
<Story
	name="Passkey not recognised"
	args={{ state: 'not-recognised' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('alert')).toHaveTextContent('That passkey is not recognised');
		await expect(canvas.getByRole('button', { name: 'Try another passkey' })).toBeEnabled();
	}}
/>
