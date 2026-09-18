<script lang="ts">
	import SegmentedControl from '../atoms/SegmentedControl.svelte';

	/**
	 * Origin `QU4bL`, gaining the owner set from `o0Tc2`. A card per account in
	 * the chooser: ownership on the left, the account in the middle, its
	 * balance on the right, an owner checkbox per other member of the
	 * household, and one level control per member who is not an owner.
	 *
	 * The owner set is what fixes the defect the single "Mine" checkbox had:
	 * each member's checkbox reports only that member's own change, so handing
	 * an account to Grace never silently drops Alan's ownership the way posting
	 * a single replacement list from one checkbox used to (ADR 0022). Ownership
	 * can never fall to zero — the database refuses it — so this control only
	 * ever adds or hands on, never orphans.
	 *
	 * Ownership and a level are mutually exclusive per member (ADR 0019): a
	 * member who is an owner does not also get a level control, because seeing
	 * a level once you already see everything is not a real choice.
	 *
	 * It carries a balance on purpose. The connecting member owns every row
	 * when the chooser opens and an owner sees their own account in full, and
	 * choosing who sees an account by its name alone, with no figure, is
	 * choosing blind.
	 */
	export type Level = 'hidden' | 'balance' | 'details';

	export interface OtherMember {
		id: string;
		name: string;
		/** Whether this member already owns the account. */
		isOwner: boolean;
		/** Meaningful only when not an owner. */
		level: Level;
	}

	interface Props {
		name: string;
		/** Bank and enough of the number to tell two accounts apart. */
		meta: string;
		/** Already formatted, with its currency and its sign. */
		balance?: string;
		initials?: string;
		/** Whether the viewing member owns this account. */
		owned: boolean;
		/** Every other member of the household. */
		others?: OtherMember[];
		/** Set when the bank offered this account for the first time on a
		 *  restore, so the member is told there is something new to choose. */
		newlyOffered?: boolean;
		/** The account is left out of wimm; the action reads "Bring back". */
		leftOut?: boolean;
		disabled?: boolean;
		onownedchange?: (owned: boolean) => void;
		/** One member's owner checkbox changed. Add or remove just that member
		 *  — never replace the whole owner set from this alone. */
		onownerchange?: (memberId: string, isOwner: boolean) => void;
		onlevelchange?: (memberId: string, level: Level) => void;
		onleaveout?: () => void;
	}

	let {
		name,
		meta,
		balance,
		initials,
		owned = $bindable(),
		others = [],
		newlyOffered = false,
		leftOut = false,
		disabled = false,
		onownedchange,
		onownerchange,
		onlevelchange,
		onleaveout
	}: Props = $props();

	const levelOptions = [
		{ value: 'hidden' as const, label: 'Nothing' },
		{ value: 'balance' as const, label: 'Balance' },
		{ value: 'details' as const, label: 'Details' }
	];

	const mark = $derived(
		initials ??
			meta
				.split(/\s+/)[0]
				?.slice(0, 2)
				.toUpperCase() ??
			''
	);

	const levelled = $derived(others.filter((m) => !m.isOwner));
</script>

<div class="row" class:disabled>
	<div class="account">
		<input
			type="checkbox"
			class="mine"
			checked={owned}
			{disabled}
			aria-label="{name} is mine"
			onchange={(event) => {
				owned = event.currentTarget.checked;
				onownedchange?.(owned);
			}}
		/>

		<span class="mark" aria-hidden="true">{mark}</span>

		<span class="identity">
			<span class="name">{name}</span>
			<span class="meta">{meta}</span>
		</span>

		{#if newlyOffered}
			<span class="badge">New</span>
		{/if}

		{#if balance}
			<span class="balance">{balance}</span>
		{/if}
	</div>

	{#if others.length > 0 || onleaveout}
		<div class="owners">
			{#if others.length > 0}
				<span class="who">Also owned by</span>
				{#each others as member (member.id)}
					<label class="co-owner">
						<input
							type="checkbox"
							checked={member.isOwner}
							{disabled}
							onchange={(event) => onownerchange?.(member.id, event.currentTarget.checked)}
						/>
						<span class="who">{member.name}</span>
					</label>
				{/each}
			{/if}
			{#if onleaveout}
				<span class="spacer"></span>
				<button type="button" class="leave-out" {disabled} onclick={onleaveout}>
					{leftOut ? 'Bring back' : 'Leave out'}
				</button>
			{/if}
		</div>
	{/if}

	{#if levelled.length > 0}
		<div class="levels">
			{#each levelled as member (member.id)}
				<div class="level">
					<span class="who" id="who-{member.id}-{name}">{member.name} sees</span>
					<SegmentedControl
						options={levelOptions}
						value={member.level}
						label="What {member.name} sees of {name}"
						{disabled}
						onchange={(level) => onlevelchange?.(member.id, level as Level)}
					/>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.row {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px;
		background: var(--color-bg-surface);
		border: 1px solid var(--color-border-subtle);
		border-radius: var(--radius-sm);
	}

	.account {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.mine {
		flex: 0 0 auto;
		inline-size: 20px;
		block-size: 20px;
		accent-color: var(--color-action-primary);
	}
	.mine:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.mark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex: 0 0 auto;
		inline-size: 32px;
		block-size: 32px;
		/* Avatar origin `Z2oYAy`: a circle on bg-inverse with inverse text,
		   instanced at 32 in a row. It was a light square, which is why the
		   marks looked nothing like the frame. */
		border-radius: var(--radius-full);
		background: var(--color-bg-inverse);
		color: var(--color-text-inverse);
		font-family: var(--type-family-body);
		font-size: 12px;
		font-weight: 600;
	}

	.identity {
		display: flex;
		flex-direction: column;
		gap: 2px;
		flex: 1 1 auto;
		min-inline-size: 0;
	}

	.name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-text-primary);
		font-family: var(--type-family-body);
		font-size: 13px;
		font-weight: 500;
	}

	.meta {
		color: var(--color-text-secondary);
		font-family: var(--type-family-mono);
		font-size: 11px;
	}

	.badge {
		flex: 0 0 auto;
		padding: 2px 8px;
		border-radius: var(--radius-pill);
		background: var(--color-feedback-info-bg);
		color: var(--color-feedback-info);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
		font-weight: 500;
	}

	.balance {
		flex: 0 0 auto;
		inline-size: 108px;
		text-align: end;
		color: var(--color-amount-neutral);
		font-family: var(--type-family-mono);
		font-size: 13px;
	}

	.owners {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.co-owner {
		display: flex;
		align-items: center;
		gap: 6px;
		cursor: pointer;
	}

	.co-owner input {
		accent-color: var(--color-action-primary);
	}

	.spacer {
		flex: 1 1 auto;
	}

	.leave-out {
		flex: 0 0 auto;
		block-size: 28px;
		padding-inline: 10px;
		border: none;
		background: transparent;
		color: var(--color-accent);
		font-family: var(--type-family-body);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
	}

	.leave-out:focus-visible {
		outline: var(--focus-ring-width) solid var(--focus-ring-color);
		outline-offset: var(--focus-ring-offset);
	}

	.levels {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.level {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.who {
		flex: 1 1 auto;
		min-inline-size: 0;
		color: var(--color-text-secondary);
		font-family: var(--type-family-body);
		font-size: var(--type-size-body-sm);
	}

	/* At Compact the level controls stack under the account, and the owner
	   checkboxes wrap: a household of three would otherwise need four columns
	   on a phone. */
	@media (max-width: 767px) {
		.level {
			flex-direction: column;
			align-items: stretch;
			gap: 4px;
		}
		.balance {
			inline-size: auto;
		}
		.owners {
			flex-wrap: wrap;
		}
	}

	/* The row's text is NOT dimmed when its controls are disabled. WCAG exempts
	   inactive controls from contrast, but these are plain text a member still
	   has to read to know which account they are looking at — and at
	   color-text-disabled on a surface they measure 2.46:1. The row reads as
	   disabled because its checkbox and its level controls genuinely are. */
</style>
