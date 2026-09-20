import { describe, expect, it } from 'vitest';
import { chartSummary, monthSection } from './overview';

describe('chartSummary', () => {
	it('names the span, both ends and the extremes with their days', () => {
		const points = [
			{ date: '22 Jun', value: 964_000, amount: '€9,640.00' },
			{ date: '14 Jul', value: 920_400, amount: '€9,204.00' },
			{ date: '2 Sep', value: 1_211_800, amount: '€12,118.00' },
			{ date: '20 Sep', value: 1_169_355, amount: '€11,693.55' }
		];

		expect(chartSummary(points)).toBe(
			'Balance from 22 Jun to 20 Sep. It starts at €9,640.00 and ends at €11,693.55. ' +
				'Highest €12,118.00 on 2 Sep. Lowest €9,204.00 on 14 Jul.'
		);
	});
});

describe('monthSection', () => {
	const eur = (minor: bigint) => ({ minor, currency: 'EUR' });
	const start = { seconds: BigInt(Date.UTC(2026, 8, 1) / 1000) };
	const today = new Date(Date.UTC(2026, 8, 20));

	it('compares against the same days of last month when it is all held', () => {
		const month = monthSection(
			{
				in: eur(245_000n),
				out: eur(181_264n),
				net: eur(63_736n),
				priorIn: eur(230_000n),
				priorOut: eur(163_024n),
				priorNet: eur(66_976n),
				monthStart: start
			},
			today
		);

		expect(month.moneyIn).toMatchObject({ change: expect.stringMatching(/150[.,]00 more/), direction: 'up' });
		expect(month.net).toMatchObject({ change: expect.stringMatching(/32[.,]40 less/), direction: 'down' });
		expect(month.net.value).toMatch(/^\+/);
		expect(month.note).toMatch(/^1 to 20 Sep, against 1 to 20 Aug\. Money moved between your own accounts is counted\.$/);
	});

	it('says where it counts from and states no comparison when a ledger starts late', () => {
		const month = monthSection(
			{
				in: eur(245_000n),
				out: eur(181_264n),
				net: eur(63_736n),
				countedFrom: { seconds: BigInt(Date.UTC(2026, 8, 19) / 1000) },
				monthStart: start
			},
			today
		);

		expect(month.moneyIn.change).toBeUndefined();
		expect(month.moneyIn.period).toBeUndefined();
		expect(month.note).toBe('Counted from 19 Sep. Money moved between your own accounts is counted.');
	});
});
