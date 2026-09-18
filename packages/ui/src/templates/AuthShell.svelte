<script lang="ts">
	import type { Snippet } from 'svelte';
	import Brand from '../atoms/Brand.svelte';
	import Icon from '../atoms/Icon.svelte';

	/**
	 * The shell every unauthenticated page sits in. Origins `kJkV1` (Wide and
	 * Ultra) and `tqiBA` (Compact).
	 *
	 * A split page, not a card on empty canvas: a 560 brand panel beside a 480
	 * card, on the same gradient the balance card uses. At Compact the panel
	 * becomes a 200 band above a 342 card, because a 390 viewport has no room
	 * for a column beside anything. That is the only structural difference
	 * between the two origins, which is why this is one component.
	 *
	 * Focus moves to the heading on load, not to the action: the member has to
	 * read who the page thinks they are before committing to anything. The
	 * action is the first tab stop after it.
	 */
	interface Props {
		heading: string;
		/**
		 * Why the member is here this time, above the heading: "You were signed
		 * out". Arrival context explains the page, so it comes before the page
		 * introduces itself.
		 */
		context?: Snippet;
		body?: Snippet;
		/**
		 * What happened when they last pressed the action, directly above it:
		 * "Your device did not save the passkey".
		 *
		 * A result belongs beside its cause. Putting it at the top would answer
		 * a question the member asked at the bottom of the card, and would push
		 * the heading down - and the heading is the identity check that
		 * `identity/passkey-enrolment` requires them to see first.
		 */
		result?: Snippet;
		/** The single action. */
		action?: Snippet;
		footnote?: Snippet;
		panelHeadline?: string;
		panelPoints?: { icon: 'circle-help' | 'shield-check' | 'smartphone'; text: string }[];
		panelFootnote?: string;
	}

	let {
		heading,
		context,
		body,
		result,
		action,
		footnote,
		panelHeadline = 'Sign in with your fingerprint, not a password.',
		panelPoints = [
			{
				icon: 'circle-help',
				text: 'Your device is the key. Nothing to choose, nothing to remember.'
			},
			{ icon: 'shield-check', text: 'The passkey stays on your device and cannot be phished.' },
			{ icon: 'smartphone', text: 'Works on the phone or laptop you already unlock every day.' }
		],
		panelFootnote = 'A household ledger'
	}: Props = $props();

	let headingElement = $state<HTMLHeadingElement | null>(null);

	$effect(() => {
		headingElement?.focus();
	});
</script>

<div class="shell">
	<aside class="panel">
		<div class="panel-brand"><Brand inverted /></div>
		<p class="panel-headline">{panelHeadline}</p>
		<ul class="panel-points">
			{#each panelPoints as point (point.text)}
				<li class="panel-point">
					<span class="point-icon"><Icon name={point.icon} size={20} /></span>
					<span>{point.text}</span>
				</li>
			{/each}
		</ul>
		<p class="panel-footnote">{panelFootnote}</p>
	</aside>

	<main class="card-area">
		<div class="card">
			{#if context}
				<div class="slot">{@render context()}</div>
			{/if}

			<h1 class="heading" bind:this={headingElement} tabindex="-1">{heading}</h1>

			{#if body}
				<p class="body">{@render body()}</p>
			{/if}

			{#if result}
				<div class="slot">{@render result()}</div>
			{/if}

			{#if action}
				<div class="action">{@render action()}</div>
			{/if}

			{#if footnote}
				<p class="footnote">{@render footnote()}</p>
			{/if}
		</div>
	</main>
</div>

<style>
	.shell {
		display: flex;
		flex-direction: column;
		min-block-size: 100dvh;
		background: var(--color-bg-canvas);
		font-family: var(--type-family-body);
		color: var(--color-text-primary);
	}

	/* Compact: a 200 band above the card, content pushed to its foot. */
	.panel {
		display: flex;
		flex-direction: column;
		justify-content: end;
		gap: var(--space-3);
		block-size: 200px;
		padding: 24px;
		background: linear-gradient(160deg, var(--gradient-brand-from), var(--gradient-brand-to));
		color: var(--color-text-on-brand);
	}

	.panel-headline {
		margin: 0;
		max-inline-size: 330px;
		font-family: var(--type-family-display);
		font-size: var(--type-size-heading-md);
		font-weight: 700;
		line-height: var(--type-line-tight);
	}

	.panel-points {
		display: none;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.panel-footnote {
		display: none;
		margin: 0;
		color: var(--color-text-on-brand-secondary);
		font-size: var(--type-size-body-sm);
	}

	.card-area {
		display: flex;
		flex: 1;
		align-items: center;
		justify-content: center;
		padding: 24px;
	}

	.card {
		display: flex;
		flex-direction: column;
		gap: var(--space-6);
		inline-size: 100%;
		max-inline-size: 342px;
		padding: var(--space-8);
		background: var(--color-bg-surface);
		border: 1px solid var(--color-border-default);
		border-radius: var(--radius-lg);
	}

	.slot {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.heading {
		margin: 0;
		font-family: var(--type-family-display);
		font-size: var(--type-size-heading-lg);
		font-weight: 700;
		letter-spacing: -0.02em;
		line-height: var(--type-line-tight);
	}

	/* No ring. The heading takes focus to place a screen reader, not to show a
	   keyboard user where they are: it is tabindex="-1", so nobody arrives here
	   by tabbing. base.css says the same for every such target; stating it here
	   too is deliberate, because a scoped `.heading:focus-visible` rule
	   outranks the attribute selector and would quietly win. */
	.heading:focus,
	.heading:focus-visible {
		outline: none;
	}

	.body {
		margin: 0;
		color: var(--color-text-secondary);
		font-size: var(--type-size-body-md);
		line-height: var(--type-line-normal);
	}

	.footnote {
		margin: 0;
		color: var(--color-text-placeholder);
		font-size: var(--type-size-body-sm);
		line-height: var(--type-line-normal);
	}

	/* Medium, Wide and Ultra: the band becomes a 560 column and the card takes
	   its 480 measure, both inside 64 of page padding. Identical across all
	   three — every screen here is Auth shell, a centred card with no
	   navigation, and there is nothing about it that changes between 768 and
	   1800. */
	@media (min-width: 768px) {
		.shell {
			flex-direction: row;
		}

		.panel {
			flex: 0 0 560px;
			block-size: auto;
			justify-content: start;
			gap: var(--space-6);
			padding: 64px;
		}

		/* Two fill spacers in the library: the lockup sits at the top, the
		   footer at the bottom, and the headline floats between them. */
		.panel-headline {
			margin-block-start: auto;
		}

		.panel-footnote {
			margin-block-start: auto;
		}

		.panel-headline {
			max-inline-size: 432px;
			font-size: var(--type-size-page-title);
		}

		.panel-points {
			display: flex;
			flex-direction: column;
			gap: var(--space-4);
			max-inline-size: 400px;
			color: var(--color-text-on-brand-secondary);
			font-size: var(--type-size-body-md);
			line-height: var(--type-line-normal);
		}

		.panel-point {
			display: flex;
			align-items: start;
			gap: var(--space-3);
		}

		.point-icon {
			flex: none;
			display: grid;
			place-items: center;
			block-size: 20px;
		}

		.panel-footnote {
			display: block;
		}

		.card-area {
			padding: 64px;
		}

		.card {
			max-inline-size: 480px;
		}
	}
</style>
