<script lang="ts">
	/**
	 * The wimm lockup: origin `vGjfy` — a 24 mark, a 9 gap, and the wordmark.
	 *
	 * The mark is the library's own two paths, `cyMbd` and `oaTkM`, read out of
	 * `ttoEX` with `includePathGeometry` and placed at the coordinates it
	 * records. Both share a scale of 9.80035 against the 16 frame, which is
	 * what the viewBox encodes, so this is the drawn mark rather than a
	 * likeness of it.
	 *
	 * `inverted` is for the brand panel, where the stock mark would be brand
	 * green on a brand-green ground: the tile takes the light colour and the
	 * glyph takes the brand.
	 */
	interface Props {
		inverted?: boolean;
		/** Hides the wordmark, leaving the mark. */
		markOnly?: boolean;
	}

	let { inverted = false, markOnly = false }: Props = $props();
</script>

<span class="brand" class:inverted>
	<span class="mark">
		<svg
			viewBox="0 0 156.8055 156.8055"
			width="16"
			height="16"
			fill="currentColor"
			aria-hidden="true"
			focusable="false"
		>
			<path
				transform="translate(0 27.1509)"
				d="M126.52777 85.27776l-0.27777 0.1389c-1.3889 0.55556-3.05557 0.83333-4.86111 0.83333-6.52779 0-10.83335-4.44444-12.36112-12.36111l-10.69444-57.77777c-1.52779-8.47222-6.25-13.19444-13.61112-13.19445-5.27777 0-12.08333 2.63889-14.30554 13.75l-10.41667 58.88889c-1.11111 5.83334-5 10.69444-11.11111 10.69444-5.83334 0-10-3.75-11.11111-11.25l-8.75001-51.94444c-2.08333-13.61111-12.36111-23.05555-26.24999-23.05555l-2.77778 0 0 14.86111 1.25 0c5.55556-0.13889 10.13889 3.19444 11.66667 8.61111 0.83333 5 9.02778 53.33333 9.16666 54.02777 2.36111 14.02777 10.69445 23.33335 26.38889 24.86112 13.47222 0.27777 23.33334-7.5 26.52778-21.25l9.44444-54.44445 8.75 49.30556c2.22221 13.75 10.13889 24.86112 25.83333 26.38889 2.63889 0.13888 5.13889-0.13889 7.22223-0.41667l0 0 0-16.80556 0.27777 0.13888z"
			/>
			<path
				transform="translate(137.6355 109.7898)"
				d="M9.58672 0c-5 0-9.44444 4.02777-9.58335 9.44443-0.13889 5.41667 4.02779 10.41667 9.58335 10.41667 5.41667 0.13889 9.58333-3.75 9.58332-9.58333 0-4.86111-3.88889-10.27777-9.58332-10.27777z"
			/>
		</svg>
	</span>
	{#if !markOnly}<span class="name">wimm</span>{/if}
</span>

<style>
	.brand {
		display: inline-flex;
		align-items: center;
		/* Literal 9 in the library: between space-2 and space-3, and optical. */
		gap: 9px;
		color: var(--color-text-primary);
	}

	.mark {
		display: grid;
		place-items: center;
		flex: none;
		inline-size: 24px;
		block-size: 24px;
		border-radius: var(--radius-sm);
		background: var(--color-action-primary);
		color: var(--color-text-inverse);
	}

	.name {
		font-family: var(--type-family-display);
		font-size: 16px;
		font-weight: 600;
		letter-spacing: -0.32px;
		line-height: var(--type-line-tight);
		color: var(--color-action-primary);
	}

	.inverted .name {
		color: var(--color-text-on-brand);
	}

	.inverted {
		color: var(--color-text-on-brand);
	}

	.inverted .mark {
		background: var(--color-text-on-brand);
		color: var(--color-action-primary);
	}
</style>
