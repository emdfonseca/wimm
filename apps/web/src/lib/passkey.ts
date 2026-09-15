/**
 * The two WebAuthn ceremonies, in the browser.
 *
 * The options and the response are JSON shapes the browser defines, so they are
 * passed through as strings rather than modelled: a type over them would
 * describe an envelope and guarantee nothing about the contents.
 */

/** The member closed or cancelled the prompt. Nothing was saved. */
export class PasskeyDismissed extends Error {}

/** The device or passkey manager cannot save a usable passkey. */
export class PasskeyUnsupported extends Error {}

export async function createPasskey(creationOptionsJson: string): Promise<string> {
	if (!('PublicKeyCredential' in window)) {
		throw new PasskeyUnsupported('this browser has no passkey support');
	}

	const options = PublicKeyCredential.parseCreationOptionsFromJSON(
		JSON.parse(creationOptionsJson)
	);

	const credential = await request(() =>
		navigator.credentials.create({ publicKey: options })
	);
	return JSON.stringify((credential as PublicKeyCredential).toJSON());
}

export async function getPasskey(requestOptionsJson: string): Promise<string> {
	if (!('PublicKeyCredential' in window)) {
		throw new PasskeyUnsupported('this browser has no passkey support');
	}

	const options = PublicKeyCredential.parseRequestOptionsFromJSON(
		JSON.parse(requestOptionsJson)
	);

	const credential = await request(() => navigator.credentials.get({ publicKey: options }));
	return JSON.stringify((credential as PublicKeyCredential).toJSON());
}

async function request(run: () => Promise<Credential | null>): Promise<Credential> {
	try {
		const credential = await run();
		if (!credential) throw new PasskeyDismissed('no credential was returned');
		return credential;
	} catch (caught) {
		if (caught instanceof PasskeyDismissed) throw caught;
		// NotAllowedError is what a browser reports both for a cancelled prompt
		// and for a timeout. Neither saved anything, and both leave the member
		// able to try again, so they are one outcome here.
		if (caught instanceof DOMException && caught.name === 'NotAllowedError') {
			throw new PasskeyDismissed(caught.message);
		}
		throw new PasskeyUnsupported(caught instanceof Error ? caught.message : String(caught));
	}
}
