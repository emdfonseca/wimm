<script lang="ts">
	import Button from '../atoms/Button.svelte';

	/**
	 * Mirrors `Dialog` (`kEly8`), over `color-scrim`, in the two shapes the
	 * canvas draws beside `J09.A / 04` and `J09.A / 06`: leaving an account out
	 * and bringing it back. Same shape, different stakes — leaving out is
	 * offered as a confirmation like any other; bringing one back is required
	 * rather than offered, because the consequence lands on a member who is
	 * not in the room, so its body always names who will see the account again
	 * and at what level.
	 *
	 * Focus is trapped, the safe action is focused first, and focus returns to
	 * whatever opened it — confirmed or dismissed.
	 */
	export interface Grantee {
		name: string;
		/** Never 'hidden': a grantee with nothing to see does not appear here. */
		level: 'balance' | 'details';
	}

	interface Props {
		open: boolean;
		mode: 'leave-out' | 'bring-back';
		accountName: string;
		/** Who will see the account again, and at what level. Meaningful only
		 *  when mode is 'bring-back'. */
		grantees?: Grantee[];
		onconfirm?: () => void;
		oncancel?: () => void;
	}

	let { open = $bindable(), mode, accountName, grantees = [], onconfirm, oncancel }: Props = $props();

	let dialog: HTMLElement | undefined = $state();
	let safeButton: HTMLButtonElement | undefined = $state();
	let opener: Element | null = null;

	$effect(() => {
		if (open) {
			opener = document.activeElement;
			queueMicrotask(() => safeButton?.focus());
		}
	});

	function close(action?: () => void) {
		open = false;
		action?.();
		if (opener instanceof HTMLElement) opener.focus();
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			event.preventDefault();
			close(oncancel);
			return;
		}
		if (event.key !== 'Tab' || !dialog) return;

		const focusable = dialog.querySelectorAll<HTMLElement>('button');
		const first = focusable[0];
		const last = focusable[focusable.length - 1];
		if (!first || !last) return;

		if (event.shiftKey && document.activeElement === first) {
			event.preventDefault();
			last.focus();
		} else if (!event.shiftKey && document.activeElement === last) {
			event.preventDefault();
			first.focus();
		}
	}

	/** "Grace will see its balance again and Alan will see its balance and its
	 *  details, which is what each of them had before you left it out." */
	const bringBackBody = $derived.by(() => {
		const phrase = (g: Grantee) =>
			g.level === 'details' ? 'its balance and its details' : 'its balance';
		if (grantees.length === 0) {
			return 'wimm starts reading it from now.';
		}
		const clauses = grantees.map(
			(g, i) => `${g.name} will see ${phrase(g)}${i === 0 ? ' again' : ''}`
		);
		const joined =
			clauses.length === 1
				? clauses[0]
				: `${clauses.slice(0, -1).join(', ')} and ${clauses[clauses.length - 1]}`;
		const had = grantees.length === 1 ? 'they had' : 'each of them had';
		return `${joined}, which is what ${had} before you left it out. wimm starts reading it from now.`;
	});
</script>

<svelte:window on:keydown={open ? onkeydown : undefined} />

{#if open}
	<div class="scrim">
		<div
			class="dialog"
			role="alertdialog"
			aria-modal="true"
			aria-labelledby="left-out-title"
			aria-describedby="left-out-message"
			bind:this={dialog}
		>
			<div class="body">
				{#if mode === 'leave-out'}
					<h2 id="left-out-title">Leave {accountName} out of wimm?</h2>
					<p id="left-out-message">
						wimm stops reading it, and it disappears for everybody else in the household. Nothing
						already read is deleted, and you can bring it back whenever you like.
					</p>
				{:else}
					<h2 id="left-out-title">Bring {accountName} back?</h2>
					<p id="left-out-message">{bringBackBody}</p>
				{/if}
			</div>

			<div class="footer">
				{#if mode === 'leave-out'}
					<Button variant="secondary" onclick={() => close(oncancel)} bind:ref={safeButton}>
						Keep it
					</Button>
					<Button variant="destructive" onclick={() => close(onconfirm)}>Leave it out</Button>
				{:else}
					<Button variant="secondary" onclick={() => close(oncancel)} bind:ref={safeButton}>
						Leave it out
					</Button>
					<Button variant="primary" onclick={() => close(onconfirm)}>Bring it back</Button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.scrim {
		position: fixed;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 16px;
		background: var(--color-scrim);
	}

	.dialog {
		display: flex;
		flex-direction: column;
		gap: 20px;
		inline-size: 100%;
		max-inline-size: 520px;
		padding: 20px;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-lg);
	}

	.body {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	h2 {
		margin: 0;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-heading-sm);
		font-weight: 600;
	}

	p {
		margin: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-md);
	}

	.footer {
		display: flex;
		gap: 12px;
		justify-content: flex-end;
	}

	@media (max-width: 767px) {
		.footer {
			flex-direction: column-reverse;
		}
	}
</style>
