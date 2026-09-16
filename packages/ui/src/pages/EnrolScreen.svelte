<script lang="ts" module>
	/**
	 * J01.A / 01 · Enrolment invitation, with J01.A / 02 · Creating passkey.
	 *
	 * Four states on one step rather than a fork, because each leaves the member
	 * here with an immediate retry.
	 */
	export type EnrolState = 'default' | 'creating' | 'not-saved' | 'dismissed';
</script>

<script lang="ts">
	import AuthShell from '../templates/AuthShell.svelte';
	import PendingButton from '../molecules/PendingButton.svelte';
	import ErrorNotice from '../molecules/ErrorNotice.svelte';
	import InfoNotice from '../molecules/InfoNotice.svelte';

	interface Props {
		/** The name the operator registered them under. */
		firstName: string;
		lastName: string;
		state?: EnrolState;
		onenrol?: () => void;
	}

	let { firstName, lastName, state = 'default', onenrol }: Props = $props();

	const fullName = $derived([firstName, lastName].filter(Boolean).join(' '));

	// The action says what it will do next, which is not the same sentence
	// after a refusal as it is the first time.
	const label = $derived(state === 'not-saved' ? 'Try again' : 'Create a passkey');

	const footnoteText = $derived(
		state === 'creating'
			? 'Your browser is asking you to confirm. This window is waiting.'
			: state === 'not-saved'
				? 'Your link still works.'
				: 'This link works once. After that, ask for a new one.'
	);
</script>

<AuthShell heading="Hello, {fullName}">
	{#snippet result()}
		{#if state === 'not-saved'}
			<ErrorNotice title="Your device did not save the passkey">
				wimm needs a passkey your device can offer you next time. Try again, or open this link on a
				phone or with a password manager that saves passkeys.
			</ErrorNotice>
		{:else if state === 'dismissed'}
			<InfoNotice title="Nothing was created">
				You closed the passkey prompt. Your link still works. Try again when you are ready.
			</InfoNotice>
		{/if}
	{/snippet}

	{#snippet body()}
		A wimm account has been set up for you. Create a passkey and you are in. There is no password to
		choose. From now on you sign in with your fingerprint, face or device PIN.
	{/snippet}

	{#snippet action()}
		<PendingButton
			{label}
			pendingLabel="Waiting for your device…"
			pending={state === 'creating'}
			onclick={onenrol}
		/>
	{/snippet}

	{#snippet footnote()}{footnoteText}{/snippet}
</AuthShell>
