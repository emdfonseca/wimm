<script lang="ts">
	/**
	 * J11.A / 01 · Settings / Appearance, and J11.B / 01 · Settings / On a
	 * touch screen.
	 *
	 * One section exists: a subsection list is for a longer or growing set of
	 * peers, and one section is neither (ADR 0005) — so this does not use
	 * `Page`'s `sections` slot, and nothing here is provisional the way the
	 * library's own Settings templates mark their six illustrative rows.
	 *
	 * One card holds both rows, a rule between them — not two cards. Each row
	 * is label, then the control pushed to the row's end by a spacer, at a
	 * fixed 264px from Medium up; at Compact the control drops below its label
	 * instead, at the row's full width.
	 *
	 * Nothing is saved as a separate step and nothing here can be wrong, so
	 * there is no saving state and no validation error to draw.
	 */
	import Page from '../templates/Page.svelte';
	import ThemeToggle, { type Theme } from '../molecules/ThemeToggle.svelte';
	import DensityControl, { type Density } from '../molecules/DensityControl.svelte';

	interface Props {
		ontheme?: (theme: Theme) => void;
		ondensity?: (density: Density) => void;
	}

	let { ontheme, ondensity }: Props = $props();
</script>

<Page title="Settings">
	<section class="appearance" aria-labelledby="appearance-heading">
		<h2 id="appearance-heading">Appearance</h2>

		<div class="row">
			<span class="label">Theme</span>
			<span class="spacer"></span>
			<div class="control">
				<ThemeToggle label="Theme" onchange={ontheme} />
			</div>
		</div>
		<p class="helper">wimm follows whatever your device is set to unless you choose here.</p>

		<div class="rule coarse-hidden"></div>

		<div class="row coarse-hidden">
			<span class="label">Rows</span>
			<span class="spacer"></span>
			<div class="control">
				<DensityControl label="Rows" onchange={ondensity} />
			</div>
		</div>
		<p class="helper coarse-hidden">
			Compact makes every row shorter, so about five more transactions fit on a laptop screen. It
			changes nothing else.
		</p>
		<!-- ADR 0004: the row is absent, not disabled, where the pointer is
		     coarse, and the card still needs to say why rather than leave a
		     gap. -->
		<p class="helper coarse-only">
			Row height is not offered here. Compact rows are smaller than a finger reliably hits, so on a
			touch screen wimm stays comfortable.
		</p>
	</section>
</Page>

<style>
	.appearance {
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 20px;
		background: var(--color-bg-surface);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-md);
	}

	h2 {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-heading-sm);
		font-weight: 600;
	}

	.row {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.label {
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
		font-weight: 500;
	}

	.control {
		inline-size: 100%;
	}

	.spacer {
		display: none;
	}

	.rule {
		block-size: 1px;
		background: var(--color-border-subtle);
	}

	.helper {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		line-height: 1.45;
	}

	.coarse-only {
		display: none;
	}

	@media (any-pointer: coarse) {
		.coarse-hidden {
			display: none;
		}
		.coarse-only {
			display: block;
		}
	}

	/* Medium and up: the row is horizontal, the control fixed at the row's
	   end rather than stretched across it. */
	@media (min-width: 768px) {
		.row {
			flex-direction: row;
			align-items: center;
			gap: 16px;
		}

		.spacer {
			display: block;
			flex: 1;
		}

		.control {
			inline-size: 264px;
		}
	}
</style>
