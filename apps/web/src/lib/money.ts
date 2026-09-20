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
export function formatMoney(
	money?: { minor: bigint | number; currency: string },
	options: { signed?: boolean; whole?: boolean } = {}
): string | undefined {
	if (!money) return undefined;

	// ISO 4217's own "no currency" code. A gateway reports it for an account
	// it cannot express as a single currency — a multi-currency wallet such as
	// PayPal, aggregated as one account — and 0 minor units alongside it is
	// not a real zero balance, so nothing is shown rather than a false one.
	if (money.currency.toUpperCase() === 'XXX') return undefined;

	const format = new Intl.NumberFormat(undefined, {
		style: 'currency',
		currency: money.currency
	});
	const digits = format.resolvedOptions().maximumFractionDigits ?? 2;

	// `signed` puts a `+` on money arriving and leaves zero bare; `whole` drops
	// the minor units, for a chart's scale where cents are noise.
	const shown = options.signed || options.whole
		? new Intl.NumberFormat(undefined, {
				style: 'currency',
				currency: money.currency,
				...(options.signed ? { signDisplay: 'exceptZero' as const } : {}),
				...(options.whole ? { maximumFractionDigits: 0, minimumFractionDigits: 0 } : {})
			})
		: format;

	return shown.format(Number(money.minor) / 10 ** digits);
}

/** True when an amount is below zero, so the sign carries direction. */
export function isNegative(money?: { minor: bigint | number }): boolean {
	return money !== undefined && Number(money.minor) < 0;
}
