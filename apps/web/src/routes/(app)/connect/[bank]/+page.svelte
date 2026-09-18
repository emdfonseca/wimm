<script lang="ts">
	import { ConsentExplainerScreen } from '@wimm/ui';
	import { resolve } from '$app/paths';

	let { data } = $props();

	let form: HTMLFormElement | undefined = $state();
</script>

<form method="POST" bind:this={form}>
	<!-- The name, not just the id: Overview's notices name the bank, and the
	     bank's return carries nothing to look it up by. -->
	<input type="hidden" name="bankName" value={data.bankName} />
	<ConsentExplainerScreen
		bankName={data.bankName}
		backHref={resolve('/(app)/connect')}
		oncontinue={() => form?.requestSubmit()}
		oncancel={() => history.back()}
	/>
</form>
