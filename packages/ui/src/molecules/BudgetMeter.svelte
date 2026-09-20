<script lang="ts">
	/**
	 * A label, a value and a track filled to a proportion: one per top
	 * merchant. The words carry everything; the track is a picture of how the
	 * value compares, so it is hidden from assistive technology.
	 */
	interface Props {
		label: string;
		/** Already formatted: `€412.60 · 9 payments`. */
		value: string;
		/** 0 to 1 of the track. */
		proportion: number;
	}

	let { label, value, proportion }: Props = $props();

	const percent = $derived(Math.min(1, Math.max(0, proportion)) * 100);
</script>

<div class="meter">
	<span class="row">
		<span class="label">{label}</span>
		<span class="value">{value}</span>
	</span>
	<span class="track" aria-hidden="true">
		<span class="fill" style:inline-size="{percent}%"></span>
	</span>
</div>

<style>
	.meter {
		display: flex;
		flex-direction: column;
		gap: 5px;
		inline-size: 100%;
		block-size: 28px;
	}

	.row {
		display: flex;
		align-items: baseline;
		gap: 12px;
	}

	.label {
		flex: 1 1 auto;
		min-inline-size: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 13px;
		font-weight: 500;
	}

	.value {
		flex: none;
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 12px;
	}

	.track {
		display: block;
		block-size: 7px;
		border-radius: var(--radius-full);
		background: var(--color-bg-subtle);
		overflow: hidden;
	}

	.fill {
		display: block;
		block-size: 100%;
		border-radius: var(--radius-full);
		background: var(--color-chart-1);
	}
</style>
