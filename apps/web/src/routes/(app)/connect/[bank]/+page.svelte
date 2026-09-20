<script lang="ts">
	import { ConsentExplainerScreen } from '@wimm/ui';
	import { resolve } from '$app/paths';

	let { data } = $props();

	let form: HTMLFormElement | undefined = $state();
</script>

<!-- A form post rather than a fetch: beginning the hand-off is a navigation,
     and the bank's page is where it ends. A hidden sibling, not a wrapper: the
     screen needs the shell's full width, which a <form> as a flex child does
     not get by default. -->
<form method="POST" bind:this={form} hidden>
	<!-- The name, not just the id: Overview's notices name the bank, and the
	     bank's return carries nothing to look it up by. -->
	<input type="hidden" name="bankName" value={data.bankName} />
</form>

<ConsentExplainerScreen
	bankName={data.bankName}
	backHref={resolve('/(app)/connect')}
	oncontinue={() => form?.requestSubmit()}
	oncancel={() => history.back()}
/>
