<script lang="ts">
	import SignedInLanding from '../templates/SignedInLanding.svelte';
	import { destinations } from '../destinations.js';
	import EmptyState from '../molecules/EmptyState.svelte';
	import Icon from '../atoms/Icon.svelte';

	/**
	 * J01.A / 03 and J02.A / 02 · Signed in. One screen, not two: the two
	 * journeys arrive at the same place, and a second copy would be a second
	 * thing to keep in step.
	 *
	 * This screen is Overview, so it says so: the shell marks nothing unless it
	 * is told which page it is on, and a nav that navigates correctly while
	 * marking the wrong item is worse than one marking nothing. The body is
	 * empty because there is no data behind it yet; a skeleton here would draw
	 * a promise this change does not make.
	 */
	interface Props {
		firstName: string;
		lastName: string;
		onsignout?: () => void;
	}

	let { firstName, lastName, onsignout }: Props = $props();

	const fullName = $derived(`${firstName} ${lastName}`);
</script>

<SignedInLanding memberName={fullName} destinations={destinations('/')} {onsignout}>
	<EmptyState title="Signed in as {fullName}" elevated>
		{#snippet icon()}
			<Icon name="user-round-check" size={24} />
		{/snippet}
		There is nothing here yet.
	</EmptyState>
</SignedInLanding>
