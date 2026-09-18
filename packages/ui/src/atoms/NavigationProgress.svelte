<script lang="ts">
	/**
	 * Origin `Mdi1X`. A 3px bar fixed to the very top of the viewport, showing
	 * that a navigation has been taken and not yet finished. One per app, in
	 * the root shell, never per screen: fixed and reserving no space, so
	 * nothing already on screen moves when it appears.
	 *
	 * Two timings are the whole contract. It waits `appearDelay` before
	 * appearing, so a fast navigation shows nothing at all; once shown it stays
	 * for at least `minimumVisible`, so a navigation that finishes just after it
	 * appeared does not produce a flash. Under reduced motion the segment fills
	 * the track and does not travel.
	 *
	 * It uses the accent, not the primary action: it reports where you are and
	 * is not something you press, the same reasoning Step indicator uses.
	 *
	 * This is the visible half only. The announcement is the other half and is
	 * composed where this is used, because it names the destination screen,
	 * which this component does not know.
	 */
	interface Props {
		/** True while a navigation is in flight. */
		active: boolean;
	}

	let { active }: Props = $props();

	const appearDelay = 150;
	const minimumVisible = 400;

	let visible = $state(false);
	let shownAt = 0;

	$effect(() => {
		let appearTimer: ReturnType<typeof setTimeout> | undefined;
		let hideTimer: ReturnType<typeof setTimeout> | undefined;

		if (active && !visible) {
			appearTimer = setTimeout(() => {
				visible = true;
				shownAt = Date.now();
			}, appearDelay);
		} else if (!active && visible) {
			const remaining = Math.max(0, minimumVisible - (Date.now() - shownAt));
			hideTimer = setTimeout(() => {
				visible = false;
			}, remaining);
		}

		return () => {
			clearTimeout(appearTimer);
			clearTimeout(hideTimer);
		};
	});
</script>

{#if visible}
	<div class="track">
		<div class="segment"></div>
	</div>
{/if}

<style>
	.track {
		position: fixed;
		inset-block-start: 0;
		inset-inline: 0;
		block-size: 3px;
		background: var(--color-bg-subtle);
		overflow: hidden;
		z-index: 100;
	}

	.segment {
		block-size: 100%;
		inline-size: 40%;
		background: var(--color-accent);
		animation: travel 1.2s var(--motion-ease-standard) infinite;
	}

	@keyframes travel {
		from {
			transform: translateX(-100%);
		}
		to {
			transform: translateX(250%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.segment {
			animation: none;
			inline-size: 100%;
		}
	}
</style>
