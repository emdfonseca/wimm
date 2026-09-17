/**
 * Minor units and an ISO 4217 code into something a person reads.
 *
 * The divisor comes from the currency, never from a hard-coded 100. JPY has no
 * minor unit and KWD has three, so dividing by 100 is wrong by a factor of a
 * hundred in one direction or a thousand in the other — the same mistake the
 * Go-side parser refuses rather than guesses at.
 *
 * `Intl` already knows the exponent, so it is asked rather than tabulated here.
 */
export function formatMoney(money?: { minor: bigint | number; currency: string }): string | undefined {
	if (!money) return undefined;

	const format = new Intl.NumberFormat(undefined, {
		style: 'currency',
		currency: money.currency
	});
	const digits = format.resolvedOptions().maximumFractionDigits ?? 2;

	return format.format(Number(money.minor) / 10 ** digits);
}

/** True when an amount is below zero, so the sign carries direction. */
export function isNegative(money?: { minor: bigint | number }): boolean {
	return money !== undefined && Number(money.minor) < 0;
}
