<script lang="ts">
	import Button from '../atoms/Button.svelte';

	/**
	 * Mirrors `Dialog` (`kEly8`), over `color-scrim`.
	 *
	 * A dialog rather than a page because it is destructive and money-bearing,
	 * and it names the bank and how many accounts go rather than asking "Are
	 * you sure?" about nothing.
	 *
	 * Focus is trapped, the safe action is focused first, and focus returns to
	 * whatever opened it — confirmed or dismissed. A member who reaches this by
	 * mistake should be able to leave it by reflex.
	 */
	interface Props {
		open: boolean;
		bankName: string;
		accountCount: number;
		onconfirm?: () => void;
		oncancel?: () => void;
	}

	let { open = $bindable(), bankName, accountCount, onconfirm, oncancel }: Props = $props();

	let dialog: HTMLElement | undefined = $state();
	/**
	 * The safe action, held directly rather than found by DOM order: reordering
	 * the footer would otherwise make Enter destroy a bank connection, and
	 * nothing would say so.
	 */
	let keepButton: HTMLButtonElement | undefined = $state();
	/** Whatever had focus before, so it can be given back. */
	let opener: Element | null = null;

	$effect(() => {
		if (open) {
			opener = document.activeElement;
			// The safe action first: confirm being focused makes Enter destroy
			// a bank connection.
			queueMicrotask(() => keepButton?.focus());
		}
	});

	function close(action?: () => void) {
		open = false;
		action?.();
		// Returned whether confirmed or dismissed: losing focus to the body
		// strands a keyboard member at the top of the page.
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
</script>

<svelte:window on:keydown={open ? onkeydown : undefined} />

{#if open}
	<div class="scrim">
		<div
			class="dialog"
			role="alertdialog"
			aria-modal="true"
			aria-labelledby="disconnect-title"
			aria-describedby="disconnect-message"
			bind:this={dialog}
		>
			<div class="body">
				<h2 id="disconnect-title">Disconnect {bankName}?</h2>
				<p id="disconnect-message">
					{accountCount}
					{accountCount === 1 ? 'account' : 'accounts'} will stop being shown, for everyone who can
					see {accountCount === 1 ? 'it' : 'them'}. wimm will stop reading from {bankName}. You can
					connect it again later.
				</p>
			</div>

			<div class="footer">
				<Button variant="secondary" onclick={() => close(oncancel)} bind:ref={keepButton}>
					Keep {bankName}
				</Button>
				<Button variant="destructive" onclick={() => close(onconfirm)}>
					Disconnect {bankName}
				</Button>
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
		max-inline-size: 420px;
		padding: 20px;
		background: var(--color-bg-elevated);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-md);
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

	@media (max-width: 599px) {
		.footer {
			flex-direction: column-reverse;
		}
	}
</style>
