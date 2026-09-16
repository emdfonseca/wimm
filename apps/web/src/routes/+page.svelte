<script lang="ts">
	import { LandingScreen } from '@wimm/ui';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let { data } = $props();

	async function signout() {
		await fetch('/api/signout', { method: 'POST' });
		await goto(resolve('/signin'), { invalidateAll: true });
	}
</script>

<LandingScreen firstName={data.firstName} lastName={data.lastName} onsignout={signout} />
