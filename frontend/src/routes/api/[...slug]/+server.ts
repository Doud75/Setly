import { error } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import type { RequestHandler } from './$types';

const BACKEND_URL = env.BACKEND_INTERNAL_URL || 'http://backend:8089/api';

const handleProxy: RequestHandler = async ({ request, params, locals, getClientAddress }) => {
	const url = new URL(request.url);
	const proxyUrl = `${BACKEND_URL}/${params.slug}${url.search}`;

	const headers = new Headers(request.headers);

	if (locals.token) {
		headers.set('Authorization', `Bearer ${locals.token}`);
	}

	if (locals.activeBandId) {
		headers.set('X-Band-ID', locals.activeBandId);
	}

	try {
		headers.set('X-Forwarded-For', getClientAddress());
	} catch {
		// Ignore if getClientAddress is not available (e.g. during prerendering)
	}

	try {
		const res = await globalThis.fetch(proxyUrl, {
			method: request.method,
			headers: headers,
			body: request.method !== 'GET' && request.method !== 'HEAD' ? request.body : null,
			duplex: 'half'
		});

		// Recopie dans une Response à headers mutables : la Response brute de fetch()
		// a des Headers immuables, et SvelteKit doit pouvoir y attacher les Set-Cookie
		// posés lors d'un refresh de token (sinon TypeError: immutable -> 500).
		const responseHeaders = new Headers(res.headers);
		responseHeaders.delete('content-encoding'); // body déjà décodé par fetch
		responseHeaders.delete('content-length'); // longueur invalide après recopie

		return new Response(res.body, {
			status: res.status,
			statusText: res.statusText,
			headers: responseHeaders
		});
	} catch (e) {
		console.error('API proxy error:', e);
		throw error(500, 'Could not connect to the backend API.');
	}
};

export const GET = handleProxy;
export const POST = handleProxy;
export const PUT = handleProxy;
export const PATCH = handleProxy;
export const DELETE = handleProxy;