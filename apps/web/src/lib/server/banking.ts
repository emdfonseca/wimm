// The banking Connect client, server side only.
//
// Same rule as identity (ADR 0001): Connect is spoken server to server, and
// `$lib/server` is SvelteKit's own guarantee that importing this from client
// code is a build error rather than a convention. The browser gets JSON from
// load functions and form actions, and never a protobuf runtime.
import { createClient, type Client } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-node';
import { BankingService } from '@wimm/contracts/banking';
import { env } from '$env/dynamic/private';

const baseUrl = env.WIMM_API_URL ?? `http://${env.WIMM_PUBLIC_ADDR ?? '127.0.0.1:9467'}/api`;

const transport = createConnectTransport({ baseUrl, httpVersion: '1.1' });

export const banking: Client<typeof BankingService> = createClient(BankingService, transport);
