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

<!-- An expired session is the page's own subject, so the heading carries it and
     there is no notice restating what the heading is for. -->
<Story
	name="Session expired"
	args={{ state: 'expired' }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('heading', { level: 1 })).toHaveTextContent('Your session ended');
		await expect(canvas.queryByRole('alert')).toBeNull();
		// No notice restating the heading. PendingButton's own live region is a
		// role="status" and is always present, so the check is that nothing
		// announces anything, not that no status element exists.
		for (const status of canvas.queryAllByRole('status')) {
			await expect(status).toHaveTextContent('');
		}
		await expect(canvas.getByText(/go back to where you were/)).toBeInTheDocument();
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
