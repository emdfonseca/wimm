import { describe, expect, it } from 'vitest';
import { formatMoney, isNegative } from './money';

/**
 * The divisor comes from the currency. Dividing by 100 unconditionally is wrong
 * by a factor of a hundred for JPY and a thousand for KWD, and it is the kind
 * of wrong that puts a plausible number on a household's screen.
 */
describe('formatMoney', () => {
	it('formats a two-decimal currency', () => {
		expect(formatMoney({ minor: 420_010n, currency: 'EUR' })).toMatch(/4[.,]200[.,]10/);
	});

	it('does not divide a currency with no minor unit', () => {
		// 1234 yen is ¥1,234 — not ¥12.34.
		const formatted = formatMoney({ minor: 1234n, currency: 'JPY' });
		expect(formatted).toMatch(/1[.,]234/);
		expect(formatted).not.toMatch(/12[.,]34/);
	});

	it('divides a three-decimal currency by a thousand', () => {
		// 1234 fils is KD 1.234.
		expect(formatMoney({ minor: 1234n, currency: 'KWD' })).toMatch(/1[.,]234/);
	});

	it('keeps the sign on an overdraft', () => {
		expect(formatMoney({ minor: -31_240n, currency: 'EUR' })).toMatch(/-|−/);
	});

	it('formats zero rather than hiding it', () => {
		expect(formatMoney({ minor: 0n, currency: 'EUR' })).toMatch(/0[.,]00/);
	});

	it('is undefined when there is no amount', () => {
		expect(formatMoney(undefined)).toBeUndefined();
	});
});

describe('isNegative', () => {
	it('reports direction from the amount', () => {
		expect(isNegative({ minor: -1n })).toBe(true);
		expect(isNegative({ minor: 0n })).toBe(false);
		expect(isNegative({ minor: 1n })).toBe(false);
		expect(isNegative(undefined)).toBe(false);
	});
});
