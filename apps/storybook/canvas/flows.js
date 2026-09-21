// Flows the design canvas can draw and play. Plain ESM because a static page
// cannot import TypeScript. Story ids derive from title and story name.

/** @type {Record<string, string>} an href the shell or a screen links to, and the story it stands for */
export const routes = {
	'/': 'pages-overview--populated',
	'/accounts': 'pages-accountsscreen--connected',
	'/transactions': 'pages-transactionsscreen--as-it-opens',
	'/settings': 'pages-settingsscreen--default'
};

/**
 * A flow is a journey a member would name. Its steps are the happy path, as the
 * routes in apps/web really run it. Its branches are what the member can meet
 * instead at a step, in parallel. A notice the member arrives to hangs under the
 * step they left, not the screen it renders: "Consent declined" is an Accounts
 * state and leaves from the consent explainer, where the member went to the bank.
 * A transition names a real callback prop, which is what play mode follows; a
 * step reached by a link has none.
 * @type {Record<string, import('./lib.js').Flow>}
 */
export const flows = {
	'enrol-a-passkey': {
		title: 'Enrol a passkey',
		steps: ['pages-enrolscreen--default', 'pages-overview--before-any-bank-connected'],
		transitions: [
			{
				from: 'pages-enrolscreen--default',
				on: 'onenrol',
				to: 'pages-overview--before-any-bank-connected'
			}
		],
		branches: [
			{
				from: 'pages-enrolscreen--default',
				outcome: 'the link had expired, been used or been replaced',
				to: 'pages-linkunusablescreen--default'
			},
			{
				from: 'pages-enrolscreen--default',
				outcome: 'the device is being asked',
				to: 'pages-enrolscreen--creating'
			},
			{
				from: 'pages-enrolscreen--default',
				outcome: 'the passkey prompt was closed',
				to: 'pages-enrolscreen--prompt-dismissed'
			},
			{
				from: 'pages-enrolscreen--default',
				outcome: 'the device did not save the passkey',
				to: 'pages-enrolscreen--passkey-not-saved'
			}
		]
	},
	'sign-in': {
		title: 'Sign in',
		steps: ['pages-signinscreen--default', 'pages-overview--populated'],
		transitions: [
			{
				from: 'pages-signinscreen--default',
				on: 'onsignin',
				to: 'pages-overview--populated'
			}
		],
		branches: [
			{
				from: 'pages-signinscreen--default',
				outcome: 'they were signed out and sent here',
				to: 'pages-signinscreen--session-expired'
			},
			{
				from: 'pages-signinscreen--default',
				outcome: 'the device is being asked',
				to: 'pages-signinscreen--signing-in'
			},
			{
				from: 'pages-signinscreen--default',
				outcome: 'the passkey prompt was closed',
				to: 'pages-signinscreen--prompt-dismissed'
			},
			{
				from: 'pages-signinscreen--default',
				outcome: 'the passkey is not one wimm knows',
				to: 'pages-signinscreen--passkey-not-recognised'
			}
		]
	},
	'connect-a-bank': {
		title: 'Connect a bank',
		steps: [
			'pages-overview--before-any-bank-connected',
			'pages-accountsscreen--empty',
			'pages-choosebankscreen--default',
			'pages-consentexplainerscreen--default',
			'pages-accountsscreen--just-connected'
		],
		transitions: [
			{
				from: 'pages-overview--before-any-bank-connected',
				on: 'ongotoaccounts',
				to: 'pages-accountsscreen--empty'
			},
			{
				from: 'pages-accountsscreen--empty',
				on: 'onconnect',
				to: 'pages-choosebankscreen--default'
			},
			{
				from: 'pages-choosebankscreen--default',
				on: 'onselect',
				to: 'pages-consentexplainerscreen--default'
			},
			{
				from: 'pages-consentexplainerscreen--default',
				on: 'oncontinue',
				to: 'pages-accountsscreen--just-connected'
			}
		],
		branches: [
			{
				from: 'pages-choosebankscreen--default',
				outcome: 'the list of banks is still loading',
				to: 'pages-choosebankscreen--loading'
			},
			{
				from: 'pages-choosebankscreen--default',
				outcome: 'the list of banks cannot be read',
				to: 'pages-choosebankscreen--list-unavailable'
			},
			{
				from: 'pages-choosebankscreen--default',
				outcome: 'typing narrows the list',
				to: 'pages-choosebankscreen--narrowed'
			},
			{
				from: 'pages-choosebankscreen--default',
				outcome: 'no bank matches the search',
				to: 'pages-choosebankscreen--no-match'
			},
			{
				from: 'pages-consentexplainerscreen--default',
				outcome: 'on a phone',
				to: 'pages-consentexplainerscreen--compact'
			},
			{
				from: 'pages-consentexplainerscreen--default',
				outcome: 'consent declined at the bank',
				to: 'pages-accountsscreen--consent-declined'
			},
			{
				from: 'pages-consentexplainerscreen--default',
				outcome: 'the bank could not be connected',
				to: 'pages-accountsscreen--a-bank-that-could-not-be-connected'
			},
			{
				from: 'pages-consentexplainerscreen--default',
				outcome: 'the bank offered no accounts',
				to: 'pages-accountsscreen--a-bank-that-offered-no-accounts'
			},
			{
				from: 'pages-consentexplainerscreen--default',
				outcome: 'the return link was already used',
				to: 'pages-accountsscreen--a-return-that-was-already-used'
			}
		]
	},
	'see-where-the-money-is': {
		title: 'See where the money is',
		steps: ['pages-overview--populated', 'pages-accountsscreen--connected'],
		transitions: [
			{
				from: 'pages-overview--populated',
				on: 'ongotoaccounts',
				to: 'pages-accountsscreen--connected'
			}
		],
		branches: [
			{
				from: 'pages-overview--populated',
				outcome: 'a day on the chart is pointed at',
				to: 'pages-overview--a-day-pointed-at'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'the chart leaves an account out',
				to: 'pages-overview--chart-does-not-cover-every-account'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'transactions only reach back to this month',
				to: 'pages-overview--ledger-starts-this-month'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'balances but no transactions',
				to: 'pages-overview--no-transaction-history'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'nothing spent yet this month',
				to: 'pages-overview--nothing-spent-yet'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'fewer than five merchants',
				to: 'pages-overview--three-merchants'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'money in two currencies',
				to: 'pages-overview--two-currencies'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'a currency holding nothing',
				to: 'pages-overview--currency-holding-nothing'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'a household of one',
				to: 'pages-overview--household-of-one'
			},
			{
				from: 'pages-overview--populated',
				outcome: 'a member who owns no account',
				to: 'pages-overview--owns-no-account'
			},
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'an account is overdrawn',
				to: 'pages-accountsscreen--an-overdrawn-account'
			},
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'an account they were given at balance level',
				to: 'pages-accountsscreen--an-account-seen-at-balance-level'
			}
		]
	},
	'read-transactions': {
		title: 'Read transactions',
		steps: [
			'pages-overview--populated',
			'pages-transactionsscreen--as-it-opens',
			'pages-transactionsscreen--an-older-page',
			'pages-transactionsscreen--the-oldest-page'
		],
		transitions: [
			{
				from: 'pages-overview--populated',
				on: 'onseeall',
				to: 'pages-transactionsscreen--as-it-opens'
			}
		],
		branches: [
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'filtered to one account',
				to: 'pages-transactionsscreen--one-account'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'on a phone',
				to: 'pages-transactionsscreen--compact'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'on a phone, filtered to one account',
				to: 'pages-transactionsscreen--compact-one-account'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'the first visit, while the banks are asked',
				to: 'pages-transactionsscreen--syncing-on-arrival'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'the banks are asked again, rows staying put',
				to: 'pages-transactionsscreen--syncing-with-rows-already-held'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'a bank is being read for the first time',
				to: 'pages-transactionsscreen--reading-for-the-first-time'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'the bank refused another read so soon',
				to: 'pages-transactionsscreen--refresh-refused'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'one bank did not answer',
				to: 'pages-transactionsscreen--one-bank-did-not-answer'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: "a bank's access has run out",
				to: 'pages-transactionsscreen--access-has-run-out'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'no bank is connected',
				to: 'pages-transactionsscreen--no-bank-connected'
			},
			{
				from: 'pages-transactionsscreen--as-it-opens',
				outcome: 'a member who owns no account',
				to: 'pages-transactionsscreen--owns-no-account'
			}
		]
	},
	'refresh-balances': {
		title: 'Refresh balances',
		steps: ['pages-accountsscreen--connected'],
		transitions: [],
		branches: [
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'balances are being read',
				to: 'pages-accountsscreen--refreshing'
			},
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'the bank refused another read so soon',
				to: 'pages-accountsscreen--refresh-refused'
			},
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'refused, and the bank did not say until when',
				to: 'pages-accountsscreen--rate-limited-with-no-retry-time'
			},
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'one bank did not answer',
				to: 'pages-accountsscreen--one-bank-not-answering'
			}
		]
	},
	'restore-access': {
		title: "Restore a bank's access",
		steps: ['pages-accountsscreen--connected-banks', 'pages-accountsscreen--access-restored'],
		transitions: [
			{
				from: 'pages-accountsscreen--connected-banks',
				on: 'onrestore',
				to: 'pages-accountsscreen--access-restored'
			}
		],
		branches: [
			{
				from: 'pages-accountsscreen--connected-banks',
				outcome: 'a refresh finds access has run out',
				to: 'pages-accountsscreen--access-run-out'
			}
		]
	},
	'disconnect-a-bank': {
		title: 'Disconnect a bank',
		steps: ['pages-accountsscreen--connected-banks', 'pages-accountsscreen--disconnected'],
		transitions: [
			{
				from: 'pages-accountsscreen--connected-banks',
				on: 'ondisconnect',
				to: 'pages-accountsscreen--disconnected'
			}
		],
		branches: [
			{
				from: 'pages-accountsscreen--connected-banks',
				outcome: 'its history stays in Transactions',
				to: 'pages-transactionsscreen--bank-disconnected'
			}
		]
	},
	'widen-consent': {
		title: "Include a bank's transactions",
		steps: [
			'pages-transactionsscreen--bank-not-sending-transactions',
			'pages-widenconsentscreen--widening',
			'pages-transactionsscreen--bank-now-included'
		],
		transitions: [
			{
				from: 'pages-widenconsentscreen--widening',
				on: 'oncontinue',
				to: 'pages-transactionsscreen--bank-now-included'
			}
		],
		branches: [
			{
				from: 'pages-widenconsentscreen--widening',
				outcome: 'on a phone',
				to: 'pages-widenconsentscreen--compact'
			}
		]
	},
	'decide-who-sees-an-account': {
		title: 'Decide who sees an account',
		steps: ['pages-accountsscreen--connected', 'pages-chooseaccountsscreen--as-it-opens'],
		transitions: [
			{
				from: 'pages-accountsscreen--connected',
				on: 'onmanage',
				to: 'pages-chooseaccountsscreen--as-it-opens'
			}
		],
		branches: [
			{
				from: 'pages-accountsscreen--connected',
				outcome: 'a bank whose accounts are not theirs to manage',
				to: 'pages-accountsscreen--a-bank-the-member-cannot-manage'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'opened again, with choices already made',
				to: 'pages-chooseaccountsscreen--reopened-with-an-existing-choice'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'one member given the balance',
				to: 'pages-chooseaccountsscreen--one-granted-at-balance'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'two members given different levels',
				to: 'pages-chooseaccountsscreen--two-members-at-different-levels'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'another member made an owner',
				to: 'pages-chooseaccountsscreen--handed-on'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: "an account made nobody's",
				to: 'pages-chooseaccountsscreen--one-disowned'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'the last owner tried to step back',
				to: 'pages-chooseaccountsscreen--the-last-owner-cannot-step-back'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'an account was left out of wimm',
				to: 'pages-accountsscreen--an-account-left-out'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'asked to confirm leaving an account out',
				to: 'pages-chooseaccountsscreen--the-leave-out-confirmation'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'asked to confirm bringing one back, and told who will see it',
				to: 'pages-chooseaccountsscreen--the-bring-back-confirmation'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'an account is being renamed',
				to: 'pages-chooseaccountsscreen--the-rename-field'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'a left-out account can be brought back',
				to: 'pages-chooseaccountsscreen--bringing-it-back'
			},
			// No route opens the chooser in restoring mode today: a restore returns
			// straight to Accounts. Whether it should is a product question.
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'a restore brought an account not seen before',
				to: 'pages-chooseaccountsscreen--restoring-with-something-new'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'a household of one',
				to: 'pages-chooseaccountsscreen--a-household-of-one'
			},
			{
				from: 'pages-chooseaccountsscreen--as-it-opens',
				outcome: 'opened by a member who owns none of them',
				to: 'pages-chooseaccountsscreen--a-member-who-owns-none-of-it'
			}
		]
	},
	'change-how-wimm-looks': {
		title: 'Change how wimm looks',
		steps: ['pages-settingsscreen--default'],
		transitions: [],
		branches: [
			{
				from: 'pages-settingsscreen--default',
				outcome: 'dark is chosen',
				to: 'pages-settingsscreen--choosing-a-theme'
			},
			{
				from: 'pages-settingsscreen--default',
				outcome: 'compact rows are chosen',
				to: 'pages-settingsscreen--choosing-row-height'
			}
		]
	}
};

/**
 * Page stories nobody has put on a flow yet. A new page story fails the check
 * until it is on a flow or here, so a state cannot go missing from the map
 * without someone deciding that. Placing one means moving it off this list. A
 * `kind-behaviour` story is never placed and never listed: it asserts something
 * and shows nothing a state does not.
 * @type {string[]}
 */
export const unplaced = [
	'pages-overview--a-month-with-unusual-payments-opened',
	'pages-overview--a-month-with-unusual-income-opened',
	'pages-overview--a-likely-yearly-payment',
	'pages-overview--a-month-with-none-opened',
	'pages-overview--the-month-so-far-opened',
	'pages-overview--without-unusual-payments',
	'pages-overview--more-goes-out-than-comes-in',
	'pages-overview--two-full-months',
	'pages-overview--a-young-account-beside-older-ones',
	'pages-overview--ledger-begins-part-way-through',
	'pages-overview--household-scope',
	'pages-overview--household-scope-with-no-history',
	'pages-overview--household-scope-with-nothing-to-set-aside',
	'pages-overview--household-scope-is-narrower',
	'pages-overview--yours-scope',
	'pages-overview--a-month-with-transfers-left-out',
	'pages-overview--a-day-a-transfer-left'
];
