<script lang="ts">
	/**
	 * Origin `W8V14`. Where a member is in a flow with a fixed, known number of
	 * steps — connecting a bank has two steps inside wimm before the hand-off,
	 * and knowing one more screen stands between them and it is the difference
	 * between reading it and clicking past it.
	 *
	 * Only for a flow whose length is known in advance: a list that grows, or a
	 * wizard whose steps depend on earlier answers, gets no count, because a
	 * number that changes under the member is worse than none. Choosing who
	 * sees an account is not a step of this flow for the same reason — it is
	 * reachable at any time from the household's accounts.
	 *
	 * Uses the accent rather than the primary action, like Navigation progress:
	 * it reports where you are and is not something you press. The track
	 * segments are 3px, the same height, so the two read as one family.
	 */
	interface Props {
		/** 1-based: the step the member is on. */
		current: number;
		/** How many steps this flow has. */
		total: number;
	}

	let { current, total }: Props = $props();
</script>

<div class="indicator">
	<div class="track" role="presentation">
		{#each { length: total } as _, index}
			<div class="segment" class:passed={index < current}></div>
		{/each}
	</div>
	<p class="label">Step {current} of {total}</p>
</div>

<style>
	.indicator {
		display: flex;
		flex-direction: column;
		gap: 8px;
		inline-size: 100%;
	}

	.track {
		display: flex;
		gap: 4px;
		block-size: 3px;
		inline-size: 100%;
	}

	.segment {
		flex: 1 1 0;
		block-size: 3px;
		background: var(--color-border-default);
	}

	.segment.passed {
		background: var(--color-accent);
	}

	.label {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: normal;
	}
</style>
