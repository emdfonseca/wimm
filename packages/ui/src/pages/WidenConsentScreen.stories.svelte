<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { expect, within } from 'storybook/test';
	import WidenConsentScreen from './WidenConsentScreen.svelte';

	const { Story } = defineMeta({
		title: 'Pages/WidenConsentScreen',
		component: WidenConsentScreen,
		tags: ['autodocs'],
		parameters: { layout: 'fullscreen', shell: '/connect' },
		args: { bankName: 'Monzo', accessEndsOn: '17 December 2026' }
	});
</script>

<Story
	name="Widening"
	tags={['kind-state']}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByRole('heading', { name: "Include Monzo's transactions", level: 1 })
		).toBeInTheDocument();
		await expect(canvas.getByText('Transactions on the accounts you own')).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Continue to Monzo' })).toBeInTheDocument();
	}}
/>

<!-- The connection is live and narrow. Nothing on this screen calls it broken,
     expired or failing: it is reading everything it was ever granted. -->
<Story
	name="NeverDescribedAsBroken"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).queryByText(/broken|expired|failed|failing/i)
		).not.toBeInTheDocument();
	}}
/>

<!-- Declining leaves the connection exactly as it was, so Not now is a link
     back rather than a destructive action. -->
<Story
	name="DecliningIsALinkBack"
	tags={['kind-behaviour']}
	play={async ({ canvasElement }) => {
		await expect(within(canvasElement).getByRole('link', { name: 'Not now' })).toHaveAttribute(
			'href',
			'/transactions'
		);
	}}
/>

<!-- The date is the bank's own limit, and the screen says so rather than
     letting it read as a wimm policy. -->
<Story
	name="AShortLivedConsent"
	tags={['kind-behaviour']}
	args={{ bankName: 'ActivoBank', accessEndsOn: '19 September 2026' }}
	play={async ({ canvasElement }) => {
		await expect(
			within(canvasElement).getByText(/ActivoBank's own limit, not a wimm setting/)
		).toBeInTheDocument();
	}}
/>

<Story
	name="Compact"
	tags={['kind-state', 'size-compact']}
	globals={{ viewport: { value: 'compact' } }}
	play={async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(
			canvas.getByRole('heading', { name: "Include Monzo's transactions" })
		).toBeInTheDocument();
		await expect(canvas.getByText('Transactions on the accounts you own')).toBeInTheDocument();
		await expect(canvas.getByRole('button', { name: 'Continue to Monzo' })).toBeVisible();
		await expect(canvas.getByRole('link', { name: 'Not now' })).toBeVisible();
	}}
/>
