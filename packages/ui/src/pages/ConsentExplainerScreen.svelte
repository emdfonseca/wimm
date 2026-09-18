<script lang="ts" module>
	/**
	 * J03.A / 02 · What wimm will see. Step 2 of 2.
	 *
	 * The last screen before the member leaves for their bank: what is read,
	 * that it stays private to them until they choose otherwise, and that
	 * they are about to leave wimm and come back.
	 */
</script>

<script lang="ts">
	import Page from '../templates/Page.svelte';
	import StepIndicator from '../molecules/StepIndicator.svelte';
	import StepActions from '../molecules/StepActions.svelte';
	import Notice from '../atoms/Notice.svelte';

	interface Props {
		bankName: string;
		/** Where "Pick another bank" goes. */
		backHref?: string;
		oncontinue?: () => void;
		oncancel?: () => void;
	}

	let { bankName, backHref = '/connect', oncontinue, oncancel }: Props = $props();

	/** Compact drops the header's back link, the "Leaving" paragraph and takes
	 *  shorter lines throughout — drawn directly rather than passed down, so
	 *  every caller does not need to know a layout breakpoint that is this
	 *  screen's concern alone. */
	let autoCompact = $state(false);

	$effect(() => {
		if (typeof window === 'undefined' || !window.matchMedia) return;
		const mq = window.matchMedia('(max-width: 767px)');
		autoCompact = mq.matches;
		const onchange = (e: MediaQueryListEvent) => (autoCompact = e.matches);
		mq.addEventListener('change', onchange);
		return () => mq.removeEventListener('change', onchange);
	});
</script>

<Page title="wimm will read your {bankName} accounts" focusHeading>
	{#snippet action()}
		{#if !autoCompact}
			<a class="back" href={backHref}>Pick another bank</a>
		{/if}
	{/snippet}

	<div class="prose">
		<StepIndicator current={2} total={2} />

		<ul class="scope">
			<li>The name of each account, and the last four digits of its number.</li>
			{#if autoCompact}
				<li>The balance of each account, read now and again on request.</li>
				<li>Nothing else. wimm never sees your {bankName} password.</li>
			{:else}
				<li>
					The balance of each account, read now and again whenever someone asks wimm to refresh.
				</li>
				<li>
					Nothing else. wimm does not see your transactions in this version, and never sees your
					{bankName} password.
				</li>
			{/if}
		</ul>

		<Notice title="Only you can see these accounts">
			{#if autoCompact}
				Nobody else sees an account until you choose who does, and how much they see.
			{:else}
				wimm serves one household, but nothing here reaches another member until you choose them.
				Per person, you decide whether they see an account's balance or its full details.
			{/if}
		</Notice>

		{#if !autoCompact}
			<p class="leaving">You will leave wimm and come back here when {bankName} is done.</p>
		{/if}

		<StepActions
			primaryLabel="Continue to {bankName}"
			lesserLabel="Cancel"
			onPrimary={oncontinue}
			onLesser={oncancel}
		/>
	</div>
</Page>

<style>
	/* Prose, so its measure is capped (ADR 0005) — the whole screen, since it
	   is nothing but the scope, the notice and the step's own actions. */
	.prose {
		display: flex;
		flex-direction: column;
		gap: 16px;
		max-inline-size: var(--layout-content-max);
	}

	.scope {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin: 0;
		padding-inline-start: 20px;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.leaving {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.back {
		color: var(--color-accent);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 500;
		text-decoration: none;
	}

	.back:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}
</style>
