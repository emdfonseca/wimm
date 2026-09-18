<script lang="ts">
	import Button from '../atoms/Button.svelte';

	/**
	 * Origin `o1elL` (row) and `prZYD` (stacked). The row of actions at the foot
	 * of a step, one arrangement for every step of every flow, so a member who
	 * has learned where the way forward is does not have to learn it again on
	 * the next screen.
	 *
	 * The lesser action comes first in reading order and the primary last —
	 * matching Dialog — but reads lesser-then-primary left to right at Medium
	 * and above; at Compact the pair stacks with the primary on top, because a
	 * column reverses which end is nearest the thumb and reading order should
	 * put the main action first. One `flex-direction` switch gets both: the DOM
	 * order never changes.
	 *
	 * The primary names what it does rather than saying "Continue" — on the
	 * step that hands the member to their bank, it names the bank.
	 *
	 * The primary is absent, not disabled, on a step whose way forward is a
	 * choice made elsewhere on the screen (`kEjpg`: choosing a bank happens by
	 * clicking a row, so its footer carries only the lesser action).
	 */
	interface Props {
		/** Absent hides the primary button entirely. */
		primaryLabel?: string;
		lesserLabel: string;
		onPrimary?: () => void;
		onLesser?: () => void;
		primaryPending?: boolean;
	}

	let { primaryLabel, lesserLabel, onPrimary, onLesser, primaryPending = false }: Props = $props();
</script>

<div class="actions">
	{#if primaryLabel}
		<Button variant="primary" block onclick={onPrimary} disabled={primaryPending}>
			{primaryLabel}
		</Button>
	{/if}
	<Button variant="secondary" block onclick={onLesser} disabled={primaryPending}>
		{lesserLabel}
	</Button>
</div>

<style>
	.actions {
		display: flex;
		flex-direction: column;
		gap: 10px;
		inline-size: 100%;
	}

	@media (min-width: 768px) {
		.actions {
			flex-direction: row-reverse;
			justify-content: flex-start;
		}

		.actions :global(.button) {
			inline-size: auto;
		}
	}
</style>
