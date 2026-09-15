// The Connect client, server side only.
//
// ADR 0001: Connect is spoken server to server. This module is imported by
// `+page.server.ts` and `+server.ts` and never by a component, which is what
// keeps the protobuf runtime out of the browser bundle. `$lib/server` is
// SvelteKit's own guarantee of that: importing it from client code is a build
// error, not a convention.
import { createClient, type Client } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { PublicService } from '@wimm/contracts/identity';
import { env } from '$env/dynamic/private';

const baseUrl = env.WIMM_API_URL ?? `http://${env.WIMM_PUBLIC_ADDR ?? '127.0.0.1:9467'}/api`;

const transport = createConnectTransport({
	baseUrl,
	httpVersion: '1.1'
});

export const identity: Client<typeof PublicService> = createClient(PublicService, transport);

/** Cookie names wimmd sets. The app relays them rather than minting its own. */
export const SESSION_COOKIE = 'wimm_session';
export const ENROLMENT_COOKIE = 'wimm_enrolment';
