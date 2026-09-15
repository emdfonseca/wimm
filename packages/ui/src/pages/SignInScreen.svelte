<script lang="ts" module>
	/**
	 * J02.A / 01 · Sign in, and J02.B, which is this same card carrying a
	 * notice rather than a page of its own.
	 *
	 * "Passkey not recognised" leaves the member here with the control still
	 * live, because offering the wrong passkey is recoverable by offering
	 * another one. It is a dead end only if they have none.
	 */
	export type SignInState = 'default' | 'signing-in' | 'dismissed' | 'expired' | 'not-recognised';
</script>

<script lang="ts">
	import AuthShell from '../templates/AuthShell.svelte';
	import PendingButton from '../molecules/PendingButton.svelte';
	import ErrorNotice from '../molecules/ErrorNotice.svelte';
	import InfoNotice from '../molecules/InfoNotice.svelte';

	interface Props {
		state?: SignInState;
		onsignin?: () => void;
	}

	let { state = 'default', onsignin }: Props = $props();

	const label = $derived(
		state === 'not-recognised' ? 'Try another passkey' : 'Sign in with a passkey'
	);
</script>

<AuthShell heading="Welcome back">
	<!-- Arrival context: why the member is on this page at all, so it reads
	     before the page introduces itself. -->
	{#snippet context()}
		{#if state === 'expired'}
			<InfoNotice title="You were signed out">
				Your session ended. Sign in again and you will go back to where you were.
			</InfoNotice>
		{/if}
	{/snippet}

	{#snippet body()}
		Sign in with the passkey you saved. Your browser will offer it. There is nothing to type.
	{/snippet}

	<!-- Results of the last attempt, directly above the action. -->
	{#snippet result()}
		{#if state === 'not-recognised'}
			<ErrorNotice title="That passkey is not recognised">
				wimm has no record of it. Ask whoever set up your account for an enrolment link.
			</ErrorNotice>
		{:else if state === 'dismissed'}
			<InfoNotice title="Nothing happened">
				You closed the passkey prompt. Try again when you are ready.
			</InfoNotice>
		{/if}
	{/snippet}

	{#snippet action()}
		<PendingButton
			{label}
			pendingLabel="Waiting for your device…"
			pending={state === 'signing-in'}
			onclick={onsignin}
		/>
	{/snippet}
</AuthShell>
