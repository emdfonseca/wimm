<script lang="ts">
	import { SignInScreen } from '@wimm/ui';
	import type { SignInState } from '@wimm/ui';
	import { getPasskey, PasskeyDismissed } from '$lib/passkey';

	let { data } = $props();

	// Seeded at initialisation, not in an effect: effects do not run during
	// server rendering, so an effect would send the expired page out saying
	// "Welcome back" and correct itself only once the client hydrated. After
	// that the page owns it, so dismissing a prompt does not bring it back.
	// svelte-ignore state_referenced_locally
	let state = $state<SignInState>(data.expired ? 'expired' : 'default');

	async function signin() {
		state = 'signing-in';
		try {
			const credentialJson = await getPasskey(data.requestOptionsJson);
			const response = await fetch('/api/signin', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ ceremonyId: data.ceremonyId, credentialJson })
			});

			// Not a page of its own: the member stays here with the control live,
			// because offering the wrong passkey is recovered by offering another.
			if (response.status === 403) {
				state = 'not-recognised';
				return;
			}
			if (!response.ok) {
				state = 'default';
				return;
			}

			const { returnPath } = await response.json();
			// A full load, not a client navigation: the session cookie has just
			// changed, so every load function has to run again against it. The
			// path came from the server's own record of the attempt and was
			// checked there, which is why it is not resolved here.
			window.location.assign(returnPath || '/');
		} catch (caught) {
			state = caught instanceof PasskeyDismissed ? 'dismissed' : 'default';
		}
	}
</script>

<SignInScreen {state} onsignin={signin} />
