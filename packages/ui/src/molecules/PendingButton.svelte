<script lang="ts">
	import Button from '../atoms/Button.svelte';

	/**
	 * Button · pending preset — "an action is underway and the browser has taken
	 * over".
	 *
	 * It wraps a Button instance and redraws none of its chrome, which is what
	 * makes it a preset rather than a component. What it adds is the behaviour
	 * the canvas could only draw the look of: the control stops being
	 * interactive, it carries its own label, and both the start and the end are
	 * announced.
	 *
	 * `aria-disabled` rather than `disabled`, so the control keeps its place in
	 * the tab order and a member who tabbed to it is not dropped somewhere else
	 * when the ceremony begins.
	 */
	interface Props {
		/** The label at rest. */
		label: string;
		/** The label while the browser has taken over. */
		pendingLabel: string;
		pending?: boolean;
		variant?: 'primary' | 'secondary' | 'ghost' | 'destructive';
		size?: 'sm' | 'md' | 'lg';
		block?: boolean;
		onclick?: () => void;
	}

	let {
		label,
		pendingLabel,
		pending = false,
		variant = 'primary',
		size = 'lg',
		block = true,
		onclick
	}: Props = $props();

	// Nothing is announced until the state has actually changed, so a page that
	// loads at rest announces nothing at all.
	let announcement = $state('');
	let previous = $state<boolean | null>(null);

	$effect(() => {
		const now = pending;
		if (previous === null) {
			previous = now;
			return;
		}
		if (previous !== now) {
			previous = now;
			announcement = now ? pendingLabel : `${label}. Ready.`;
		}
	});

	function guard() {
		if (!pending) onclick?.();
	}
</script>

<Button
	{variant}
	{size}
	{block}
	onclick={guard}
	aria-disabled={pending ? 'true' : undefined}
	aria-busy={pending ? 'true' : undefined}
>
	{pending ? pendingLabel : label}
</Button>

<p class="status" role="status" aria-live="polite">{announcement}</p>

<style>
	/* Visually hidden, still announced. `display: none` would take it out of the
	   accessibility tree along with the announcement. */
	.status {
		position: absolute;
		inline-size: 1px;
		block-size: 1px;
		margin: -1px;
		padding: 0;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
		border: 0;
	}
</style>
