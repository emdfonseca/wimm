<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, fn, userEvent, within } from 'storybook/test';
	import DisconnectBankDialog from './DisconnectBankDialog.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/DisconnectBankDialog',
		component: DisconnectBankDialog,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen' },
		args: { open: true, bankName: 'Montepio', accountCount: 3, onconfirm: fn(), oncancel: fn() }
	});
</script>

<!-- It names the bank and how many accounts go, rather than asking "Are you
     sure?" about nothing. -->
<Story
	name="Default"
	play={async ({ canvasElement }) => {
		const canvas = within(document.body);
		await expect(canvas.getByRole('alertdialog')).toHaveAccessibleName('Disconnect Montepio?');
		await expect(canvas.getByText(/3 accounts will stop being shown/)).toBeInTheDocument();
	}}
/>

<Story name="One account" args={{ accountCount: 1 }} />

<!-- Both buttons name the bank: "Cancel" and "Confirm" beside each other tell
     a member nothing about which does what. -->
<Story
	name="The buttons say what they do"
	play={async () => {
		const canvas = within(document.body);
		await expect(canvas.getByRole('button', { name: 'Keep Montepio' })).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Disconnect Montepio' })).toBeInTheDocument();
	}}
/>

<!-- The safe action is focused first: confirm being focused would make Enter
     destroy a bank connection. -->
<Story
	name="The safe action is focused"
	play={async () => {
		const canvas = within(document.body);
		await expect(canvas.getByRole('button', { name: 'Keep Montepio' })).toHaveFocus();
	}}
/>

<Story
	name="Confirming"
	play={async ({ args }) => {
		await userEvent.click(within(document.body).getByRole('button', { name: 'Disconnect Montepio' }));
		await expect(args.onconfirm).toHaveBeenCalledOnce();
	}}
/>

<Story
	name="Changing their mind"
	play={async ({ args }) => {
		await userEvent.click(within(document.body).getByRole('button', { name: 'Keep Montepio' }));
		await expect(args.oncancel).toHaveBeenCalledOnce();
		await expect(args.onconfirm).not.toHaveBeenCalled();
	}}
/>

<!-- Escape dismisses: a member who reaches this by mistake should be able to
     leave by reflex. -->
<Story
	name="Escape dismisses"
	play={async ({ args }) => {
		await userEvent.keyboard('{Escape}');
		await expect(args.oncancel).toHaveBeenCalledOnce();
		await expect(args.onconfirm).not.toHaveBeenCalled();
	}}
/>

<!-- Focus is trapped: tabbing past the last control comes back to the first
     rather than escaping to the page behind. -->
<Story
	name="Focus is trapped"
	play={async () => {
		const canvas = within(document.body);
		const keep = canvas.getByRole('button', { name: 'Keep Montepio' });
		const disconnect = canvas.getByRole('button', { name: 'Disconnect Montepio' });

		await expect(keep).toHaveFocus();
		await userEvent.tab();
		await expect(disconnect).toHaveFocus();
		await userEvent.tab();
		await expect(keep).toHaveFocus();
		await userEvent.tab({ shift: true });
		await expect(disconnect).toHaveFocus();
	}}
/>

<Story name="Closed" args={{ open: false }} />

<!-- Ending wimm's access and destroying the record of what it read are
     different decisions. The confirmation says which one it is, because that is
     what a member is weighing. -->
<Story
	name="TheRecordIsKept"
	args={{ open: true }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/transactions already read stay under Transactions/)
		).toBeInTheDocument();
	}}
/>
