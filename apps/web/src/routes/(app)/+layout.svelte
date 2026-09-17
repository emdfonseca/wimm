<script lang="ts">
	import { SignedInLanding, ThemeToggle, destinations } from '@wimm/ui';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';

	let { data, children } = $props();

	// Which destination this is. Only the route knows, and the shell marks
	// nothing without being told — so this is where the sidebar and the bottom
	// bar learn where the member is.
	const here = $derived(destinations(page.url.pathname));

	async function signout() {
		await fetch('/api/signout', { method: 'POST' });
		await goto(resolve('/signin'), { invalidateAll: true });
	}
</script>

<!-- The shell wraps every signed-in screen from here, so a screen added later
     cannot render outside it. -->
<SignedInLanding
	memberName={`${data.firstName} ${data.lastName}`}
	destinations={here}
	onsignout={signout}
>
	{@render children()}

	<!-- In the shell rather than on a screen: it belongs to the app, not to
	     whatever page happens to be open. -->
	<div class="appearance">
		<span class="appearance-label">Appearance</span>
		<ThemeToggle />
	</div>
</SignedInLanding>

<style>
	.appearance {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-block-start: 32px;
		padding-block-start: 16px;
		border-block-start: 1px solid var(--color-border-subtle);
	}

	.appearance-label {
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}
</style>
