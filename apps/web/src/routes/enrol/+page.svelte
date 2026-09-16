<script lang="ts">
	import { EnrolScreen } from '@wimm/ui';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { createPasskey, PasskeyDismissed, PasskeyUnsupported } from '$lib/passkey';
	import type { EnrolState } from '@wimm/ui';

	/**
	 * Route wiring only. The screen is a pure component in `@wimm/ui`: it takes
	 * data as props and emits intent as a callback, which is what lets it be a
	 * story with nothing mocked.
	 */
	let { data } = $props();

	let state = $state<EnrolState>('default');

	async function enrol() {
		state = 'creating';
		try {
			const credentialJson = await createPasskey(data.creationOptionsJson);
			const response = await fetch('/api/enrolment', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ ceremonyId: data.ceremonyId, credentialJson })
			});

			if (!response.ok) {
				state = 'not-saved';
				return;
			}
			await goto(resolve('/'), { invalidateAll: true });
		} catch (caught) {
			// A closed prompt changed nothing, and says so politely. A device that
			// cannot save a discoverable credential is an error the member has to
			// act on.
			state = caught instanceof PasskeyDismissed ? 'dismissed' : 'not-saved';
			if (caught instanceof PasskeyUnsupported) state = 'not-saved';
		}
	}
</script>

<EnrolScreen firstName={data.firstName} lastName={data.lastName} {state} onenrol={enrol} />
