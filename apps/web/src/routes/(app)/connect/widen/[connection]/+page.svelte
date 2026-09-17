<script lang="ts">
	import { WidenConsentScreen } from '@wimm/ui';

	let { data } = $props();

	let form: HTMLFormElement;
</script>

<!-- A form post rather than a fetch: beginning the hand-off is a navigation,
     and the bank's page is where it ends. -->
<form method="POST" bind:this={form} hidden>
	<input type="hidden" name="bankName" value={data.bankName} />
</form>

<WidenConsentScreen
	bankName={data.bankName}
	accessEndsOn={data.accessEndsOn}
	oncontinue={() => form.requestSubmit()}
/>
