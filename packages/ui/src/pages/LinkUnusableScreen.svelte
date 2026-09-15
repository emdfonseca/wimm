<script lang="ts">
	import AuthShell from '../templates/AuthShell.svelte';

	/**
	 * J01.B / 01 · Link unusable. The only branch in either journey: the only
	 * place the goal cannot be completed and the recovery leaves the product.
	 *
	 * Expired, already used, replaced and never issued all render this, and
	 * nothing says which. No notice and no action: the card is the message, and
	 * a control that cannot work is worse than none.
	 */
</script>

<AuthShell heading="This link cannot be used">
	{#snippet body()}
		It may have expired, already been used, or been replaced by a newer one. Ask whoever set up your
		account for a new link.
	{/snippet}
</AuthShell>
