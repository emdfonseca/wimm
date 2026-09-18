<script lang="ts">
	import Button from '../atoms/Button.svelte';

	/**
	 * Origin `FguRX`. An owner giving an account the name the household
	 * actually uses for it — a bank that names three accounts the same thing,
	 * or renames its products, is the case this exists for.
	 *
	 * The bank's own name stays beneath the field rather than being replaced,
	 * so the record of what the bank said survives and a member can tell which
	 * account they are looking at while renaming it. Clearing the field
	 * returns the account to the bank's name; it does not leave it blank.
	 *
	 * Everybody who may see the account sees the household's name, including a
	 * member granted only the balance — it reveals nothing the bank's name did
	 * not.
	 */
	interface Props {
		/** The household's name, empty when nobody has set one. */
		householdName: string;
		/** Which bank, e.g. "Monzo". */
		bankName: string;
		/** The bank's own name for this account, e.g. "Current Account" —
		 *  always shown beneath the field, and the placeholder when the
		 *  household has not set one. */
		bankAccountName: string;
		onSave?: (householdName: string) => void;
	}

	let { householdName, bankName, bankAccountName, onSave }: Props = $props();

	// A local buffer, deliberately not re-synced when householdName changes
	// underneath: it is the draft the member is editing, not a mirror of props.
	let draft = $state(householdName);

	function save() {
		onSave?.(draft.trim());
	}

	function clear() {
		draft = '';
		onSave?.('');
	}
</script>

<div class="account-name">
	<label class="label" for="household-name">What we call this account</label>
	<div class="row">
		<input
			id="household-name"
			class="field"
			type="text"
			bind:value={draft}
			placeholder={bankAccountName}
		/>
		<Button variant="secondary" onclick={save}>Save</Button>
		{#if draft}
			<Button variant="ghost" onclick={clear}>Clear</Button>
		{/if}
	</div>
	<p class="helper">{bankName} calls it {bankAccountName}. Clear this to go back to that.</p>
</div>

<style>
	.account-name {
		display: flex;
		flex-direction: column;
		gap: 6px;
		inline-size: 100%;
	}

	.label {
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 600;
	}

	.row {
		display: flex;
		gap: 8px;
		inline-size: 100%;
	}

	.field {
		flex: 1 1 0;
		min-inline-size: 0;
		block-size: 36px;
		padding-inline: 12px;
		border: 1px solid var(--color-control-border);
		border-radius: var(--radius-control);
		background: var(--color-bg-surface);
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.field:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.helper {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: normal;
	}
</style>
