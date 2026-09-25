import type { LedgerFilterValues } from '@wimm/ui';

/**
 * The ledger's filters as the address carries them. The address is the only
 * place a filter lives, so paging, Back and a saved link all keep it with no
 * client state.
 */

/** The longest search wimmd reads, and the search field's maxlength. */
export const MAX_SEARCH = 100;

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
const MONTH = /^\d{4}-(0[1-9]|1[0-2])$/;

const usable: Record<keyof LedgerFilterValues, (value: string) => boolean> = {
	account: (v) => UUID.test(v),
	q: (v) => v.trim() !== '' && v.length <= MAX_SEARCH,
	month: (v) => MONTH.test(v),
	direction: (v) => v === 'in' || v === 'out'
};

const KEYS = Object.keys(usable) as (keyof LedgerFilterValues)[];

/**
 * The filters in the address, and the names of any present that wimm cannot
 * use. An empty value is one of those: it says nothing, and the address should
 * say what is on screen.
 */
export function readFilters(params: URLSearchParams): {
	filters: LedgerFilterValues;
	dropped: string[];
} {
	const filters: LedgerFilterValues = { account: '', q: '', month: '', direction: '' };
	const dropped: string[] = [];
	for (const key of KEYS) {
		const value = params.get(key);
		if (value === null) continue;
		if (usable[key](value)) Object.assign(filters, { [key]: value });
		else dropped.push(key);
	}
	return { filters, dropped };
}

/** The filters in force as address parameters, in a fixed order. */
export function filterParams(filters: LedgerFilterValues): URLSearchParams {
	const params = new URLSearchParams();
	for (const key of KEYS) if (filters[key]) params.set(key, filters[key]);
	return params;
}

/**
 * Where a filter change goes. No cursor, oldest or page survives it, which is
 * what starts reading again at the newest match.
 */
export function filterHref(filters: LedgerFilterValues): string {
	return `/transactions${filterQuery(filters)}`;
}

/** The query string of filterHref, `?…` or empty, for a caller resolving the
 *  path itself. */
export function filterQuery(filters: LedgerFilterValues): string {
	const query = filterParams(filters).toString();
	return query ? `?${query}` : '';
}

/** The same address with the named parameters removed. */
export function without(url: URL, keys: string[]): string {
	const params = new URLSearchParams(url.searchParams);
	for (const key of keys) params.delete(key);
	const query = params.toString();
	return query ? `${url.pathname}?${query}` : url.pathname;
}
