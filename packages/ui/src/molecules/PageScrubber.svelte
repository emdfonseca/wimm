<script lang="ts">
	import { fade } from 'svelte/transition';
	import Icon from '../atoms/Icon.svelte';

	/**
	 * Origin `W8ig0`, reworked as a timeline over real pages rather than
	 * calendar months. A month is not a place a member can reliably land: a
	 * household reading two transactions a day fills one 50-row page with
	 * four months of them, and a scrubber built on months lets a member click
	 * four different destinations that are, underneath, the same page. A page
	 * here is exactly what the ledger's own page index found — the same
	 * cursor, read the same way — so choosing one always lands on what it
	 * says, never on a neighbour that happened to share its page.
	 *
	 * A row of chip buttons read as a filter bar, not as time — this is a
	 * track instead: one dot per real page, oldest to newest, left to right.
	 * Moving the pointer along it previews the page under the cursor — the
	 * dot grows, the rest ease back down, and a label floats above it —
	 * before anything is chosen. Choosing still takes a click or a key,
	 * exactly as a chip did.
	 *
	 * A page is a place to start reading, not a filter: choosing one lands on
	 * that page, and there is no way to clear it because nothing was cut out
	 * of the list.
	 *
	 * Every dot is still a real link where one is given, exactly as a chip
	 * was: Back returns here, the page is bookmarkable, and each carries its
	 * own span as its accessible name — visible on hover and focus instead
	 * of always.
	 */
	export interface ScrubberPage {
		/** Unique per page — the cursor it names, opaque to this component. */
		key: string;
		/** The span of dates this page holds, e.g. "17 Sep to 9 Mar, 2026". */
		label: string;
		/** Where choosing this page goes. A real link, like Newer/Older
		 *  already are, so Back returns here and the page is bookmarkable. */
		href?: string;
	}

	interface Props {
		pages: ScrubberPage[];
		/** The page a member is actually on, by its key — or null on a page
		 *  reached some other way (paging one at a time). */
		current?: string | null;
		/** Where Newest goes. Absent means this is already the newest page. */
		newestHref?: string;
		/** Where Oldest goes. Absent means this is already the oldest page. */
		oldestHref?: string;
		onSelectPage?: (page: ScrubberPage) => void;
		onNewest?: () => void;
		onOldest?: () => void;
	}

	let {
		pages,
		current = null,
		newestHref,
		oldestHref,
		onSelectPage,
		onNewest,
		onOldest
	}: Props = $props();

	// Oldest first, left to right — flying forward through time reads the same
	// direction the page itself does (newest content at the top, oldest at
	// the bottom, time running top-to-bottom and so left-to-right here too).
	const ordered = $derived([...pages].reverse());

	/** The page under the pointer, previewed but not yet chosen. */
	let hovered = $state<string | null>(null);
	/** Fixed-position coordinates for the floating label. `position: fixed`
	 *  so it escapes the track's own `overflow-x: auto` — a descendant
	 *  positioned only relative to the track is clipped by it the moment the
	 *  track can scroll, flag included. */
	let flagX = $state(0);
	let flagY = $state(0);
	let track = $state<HTMLElement | null>(null);

	// Y never moves: it comes from the track, which never scales, rather than
	// from whichever dot is currently enlarged — reading it off the hovered
	// dot's own box made the flag hop up and down by however much that dot
	// happened to be scaled, which reads as noise rather than motion.
	function flagAnchorY(): number {
		return track?.getBoundingClientRect().top ?? 0;
	}

	function previewNearest(clientX: number) {
		if (!track || ordered.length === 0) return;
		const dots = Array.from(track.querySelectorAll<HTMLElement>('[data-dot]'));
		let closest = 0;
		let closestDistance = Infinity;
		dots.forEach((dot, i) => {
			const rect = dot.getBoundingClientRect();
			const distance = Math.abs(clientX - (rect.left + rect.width / 2));
			if (distance < closestDistance) {
				closestDistance = distance;
				closest = i;
			}
		});
		hovered = ordered[closest]!.key;
		const rect = dots[closest]?.getBoundingClientRect();
		if (rect) flagX = rect.left + rect.width / 2;
		flagY = flagAnchorY();
	}

	function previewFocused(e: FocusEvent, page: ScrubberPage) {
		hovered = page.key;
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		flagX = rect.left + rect.width / 2;
		flagY = flagAnchorY();
	}

	const hoveredPage = $derived(ordered.find((p) => p.key === hovered) ?? null);

	/** What the track is showing without hovering or focusing anything — the
	 *  page a member is actually on, stated plainly rather than left to a
	 *  slightly bigger dot to communicate on its own. */
	const currentLabel = $derived(
		current ? (ordered.find((p) => p.key === current)?.label ?? null) : null
	);
</script>

<div class="scrubber">
	{#if !hoveredPage && currentLabel}
		<span class="current">{currentLabel}</span>
	{/if}

	{#if oldestHref}
		<a class="jump" href={oldestHref} aria-label="Oldest"><Icon name="chevron-left" size={14} /></a>
	{:else if onOldest}
		<button type="button" class="jump" onclick={onOldest} aria-label="Oldest">
			<Icon name="chevron-left" size={14} />
		</button>
	{:else}
		<!-- Present, disabled: the track's own oldest end, and there is
		     nowhere further back to go. Removing the arrow moves everything
		     else beside it, which reads as the layout shifting rather than as
		     an edge reached. -->
		<button type="button" class="jump" disabled aria-label="Oldest">
			<Icon name="chevron-left" size={14} />
		</button>
	{/if}

	<div
		class="track"
		bind:this={track}
		role="group"
		aria-label="Pages"
		onpointermove={(e) => previewNearest(e.clientX)}
		onpointerleave={() => (hovered = null)}
	>
		<span class="line" aria-hidden="true"></span>
		{#each ordered as page (page.key)}
			{@const selected = current !== null && current === page.key}
			{@const preview = hovered === page.key}
			{#if page.href}
				<a
					data-dot
					class="dot"
					class:selected
					class:preview
					aria-current={selected ? 'true' : undefined}
					href={page.href}
					onclick={() => onSelectPage?.(page)}
					onfocus={(e) => previewFocused(e, page)}
					onblur={() => (hovered = null)}
				>
					<span class="sr-only">{page.label}</span>
				</a>
			{:else}
				<button
					data-dot
					type="button"
					class="dot"
					class:selected
					class:preview
					aria-current={selected ? 'true' : undefined}
					onclick={() => onSelectPage?.(page)}
					onfocus={(e) => previewFocused(e, page)}
					onblur={() => (hovered = null)}
				>
					<span class="sr-only">{page.label}</span>
				</button>
			{/if}
		{/each}
	</div>

	{#if hoveredPage}
		<span
			class="flag-anchor"
			aria-hidden="true"
			style:left="{flagX}px"
			style:top="{flagY}px"
			in:fade={{ duration: 100 }}
			out:fade={{ duration: 100 }}
		>
			<span class="flag">{hoveredPage.label}</span>
		</span>
	{/if}

	{#if newestHref}
		<a class="jump" href={newestHref} aria-label="Newest"><Icon name="chevron-right" size={14} /></a>
	{:else if onNewest}
		<button type="button" class="jump" onclick={onNewest} aria-label="Newest">
			<Icon name="chevron-right" size={14} />
		</button>
	{:else}
		<button type="button" class="jump" disabled aria-label="Newest">
			<Icon name="chevron-right" size={14} />
		</button>
	{/if}
</div>

<style>
	.scrubber {
		display: flex;
		align-items: center;
		align-self: flex-end;
		gap: 8px;
		block-size: 44px;
		inline-size: fit-content;
		max-inline-size: 100%;
		padding-inline: 8px;
	}

	.current {
		flex: none;
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 12px;
		white-space: nowrap;
	}

	.track {
		position: relative;
		display: flex;
		align-items: center;
		gap: 22px;
		flex: 1 1 0;
		min-inline-size: 0;
		overflow-x: auto;
		/* Inline padding, not just block: the first and last dot scale up
		   around their own centre exactly like every other one, and with no
		   room either side of them that growth was clipped by the scroll
		   container's own edge. */
		padding: 12px 10px;
		scrollbar-width: none;
	}

	.track::-webkit-scrollbar {
		display: none;
	}

	.line {
		position: absolute;
		inset-inline: 4px;
		block-size: 2px;
		background: var(--color-border-subtle);
		border-radius: var(--radius-full);
		pointer-events: none;
	}

	.dot {
		position: relative;
		flex: none;
		inline-size: 8px;
		block-size: 8px;
		border: none;
		border-radius: var(--radius-full);
		background: var(--color-text-secondary);
		cursor: pointer;
		transition:
			transform 220ms cubic-bezier(0.16, 1, 0.3, 1),
			background 220ms ease;
	}

	.dot:hover,
	.dot.preview {
		transform: scale(1.9);
		background: var(--color-accent);
	}

	.dot.selected {
		background: var(--color-accent);
		transform: scale(1.4);
	}

	.dot.selected.preview,
	.dot.selected:hover {
		transform: scale(1.9);
	}

	.dot:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: 4px;
	}

	@media (prefers-reduced-motion: reduce) {
		.dot {
			transition: none;
		}
	}

	/* Fixed, not relative to the track: a descendant positioned only relative
	   to its scrolling ancestor is clipped by it the moment that ancestor's
	   overflow can scroll, floating label included. Left/top transition, so
	   the anchor glides from page to page rather than jumping — the flying
	   the pointer does across the track, echoed in the label that follows it. */
	.flag-anchor {
		position: fixed;
		transform: translate(-50%, calc(-100% - 10px));
		pointer-events: none;
		z-index: 1;
		transition: left 140ms cubic-bezier(0.16, 1, 0.3, 1);
	}

	.flag {
		position: relative;
		display: block;
		padding: 6px 10px;
		border-radius: var(--radius-sm);
		background: var(--color-bg-inverse);
		color: var(--color-text-inverse);
		box-shadow: 0 4px 12px rgb(0 0 0 / 0.24);
		font-family: var(--type-family-mono);
		font-size: 12px;
		font-weight: 500;
		white-space: nowrap;
	}

	.flag::after {
		content: '';
		position: absolute;
		inset-block-start: 100%;
		inset-inline-start: 50%;
		transform: translateX(-50%);
		border: 5px solid transparent;
		border-block-start-color: var(--color-bg-inverse);
	}

	@media (prefers-reduced-motion: reduce) {
		.flag-anchor {
			transition: none;
		}
	}

	.sr-only {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
	}

	.jump {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		inline-size: 28px;
		block-size: 28px;
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-full);
		background: var(--color-action-secondary);
		color: var(--color-text-primary);
		text-decoration: none;
	}

	.jump:hover {
		background: var(--color-action-secondary-hover);
	}

	.jump:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.jump:disabled {
		color: var(--color-text-disabled);
		cursor: not-allowed;
	}

	.jump:disabled:hover {
		background: var(--color-action-secondary);
	}
</style>
