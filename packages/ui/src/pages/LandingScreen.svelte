<script lang="ts">
	import SignedInLanding from '../templates/SignedInLanding.svelte';
	import EmptyState from '../molecules/EmptyState.svelte';
	import Icon from '../atoms/Icon.svelte';

	/**
	 * J01.A / 03 and J02.A / 02 · Signed in. One screen, not two: the two
	 * journeys arrive at the same place, and a second copy would be a second
	 * thing to keep in step.
	 *
	 * Overview is the only destination that exists, so it is the only one
	 * enabled, and it is marked current. The body is empty because there is no
	 * data behind it yet; a skeleton here would draw a promise this change does
	 * not make.
	 */
	interface Props {
		firstName: string;
		lastName: string;
		onsignout?: () => void;
	}

	let { firstName, lastName, onsignout }: Props = $props();

	const fullName = $derived(`${firstName} ${lastName}`);
</script>

<SignedInLanding memberName={fullName} {onsignout}>
	<EmptyState title="Signed in as {fullName}" elevated>
		{#snippet icon()}
			<Icon name="user-round-check" size={24} />
		{/snippet}
		There is nothing here yet.
	</EmptyState>
</SignedInLanding>
