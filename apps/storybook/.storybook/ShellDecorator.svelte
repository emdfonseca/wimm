<script lang="ts">
	import type { Snippet } from 'svelte';
	import { SignedInLanding, destinations } from '@wimm/ui';

	/**
	 * The signed-in shell around a screen, as `apps/web`'s `(app)/+layout.svelte`
	 * wraps every signed-in route. A story opts in with `parameters.shell`, the
	 * path the shell should mark as current.
	 */
	let { current, children }: { current: string; children?: Snippet } = $props();
</script>

<SignedInLanding memberName="Ana Costa" destinations={destinations(current)}>
	{@render children?.()}
</SignedInLanding>
