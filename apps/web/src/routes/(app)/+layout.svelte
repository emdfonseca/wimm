<script lang="ts">
	import { SignedInLanding, destinations } from '@wimm/ui';
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
</SignedInLanding>
