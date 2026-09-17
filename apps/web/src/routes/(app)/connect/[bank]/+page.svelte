<script lang="ts">
	import { ConsentExplainerScreen } from '@wimm/ui';

	let { data } = $props();

	/**
	 * Formatted here rather than on the server so the member reads it in their
	 * own locale. The value is the bank's maximum, which wimmd asked for in
	 * full.
	 */
	const endsOn = $derived(
		new Date(Date.now() + data.maxConsentSeconds * 1000).toLocaleDateString(undefined, {
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		})
	);

	/** A day or less is worth naming plainly: it makes restoring routine. */
	const shortLived = $derived(data.maxConsentSeconds <= 48 * 60 * 60);

	let form: HTMLFormElement | undefined = $state();
</script>

<form method="POST" bind:this={form}>
	<!-- The name, not just the id: Overview's notices name the bank, and the
	     bank's return carries nothing to look it up by. -->
	<input type="hidden" name="bankName" value={data.bankName} />
	<ConsentExplainerScreen
		bankName={data.bankName}
		accessEndsOn={endsOn}
		{shortLived}
		oncontinue={() => form?.requestSubmit()}
		oncancel={() => history.back()}
	/>
</form>
