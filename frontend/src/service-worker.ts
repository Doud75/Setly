/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />

import { cachePut } from './lib/sw/cacheUtils';

declare const self: ServiceWorkerGlobalScope;

declare global {
	interface ServiceWorkerGlobalScope {
		__WB_MANIFEST: Array<{ url: string; revision: string | null }>;
	}
}

const STATIC_CACHE = 'static-v1';
const PAGES_CACHE = 'pages-v2';
const API_CACHE = 'api-v1';
const OFFLINE_URL = '/offline';

const ALL_CACHES = [STATIC_CACHE, PAGES_CACHE, API_CACHE];

self.addEventListener('install', (event: ExtendableEvent) => {
	event.waitUntil(
		caches
			.open(STATIC_CACHE)
			.then(async (cache) => {
				// Pré-cacher la page offline ET ses données SvelteKit
				await Promise.allSettled([
					cache.add(OFFLINE_URL),
					cache.add('/offline/__data.json')
				]).catch((err) => console.warn('[SW] Failed to cache offline assets:', err));
				const urls = [...new Set(self.__WB_MANIFEST.map((e) => e.url))];
				await Promise.allSettled(
					urls.map((url) =>
						cache.add(url).catch((err) => console.warn('[SW] Failed to cache:', url, err))
					)
				);
			})
			.then(async () => {
				const pagesCache = await caches.open(PAGES_CACHE);
				await Promise.allSettled(
					['/', '/login', '/signup'].map((url) =>
						fetch(url, { redirect: 'manual' })
							.then((res) => {
								// Ne jamais cacher une redirection (ex: `/` → `/login` quand
								// déconnecté), sinon on sert la page de login sous l'URL `/`.
								if (res.ok && !res.redirected && res.type !== 'opaqueredirect') {
									return pagesCache.put(url, res);
								}
							})
							.catch((err) => console.warn('[SW] Failed to cache auth page:', url, err))
					)
				);
			})
			.then(() => {
				return self.skipWaiting();
			})
	);
});

self.addEventListener('activate', (event: ExtendableEvent) => {
	event.waitUntil(
		caches
			.keys()
			.then((keys) =>
				Promise.all(keys.filter((k) => !ALL_CACHES.includes(k)).map((k) => caches.delete(k)))
			)
			.then(() => {
				return self.clients.claim();
			})
	);
});

self.addEventListener('fetch', (event: FetchEvent) => {
	const { request } = event;
	if (request.method !== 'GET') return;

	const url = new URL(request.url);
	if (url.origin !== self.location.origin) return;

	const path = url.pathname;

	if (path === '/_app/version.json') return;

	if (path.startsWith('/_app/')) {
		event.respondWith(cacheFirst(request));
		return;
	}

	if (path.startsWith('/api/auth') || path === '/logout') return;

	if (request.mode === 'navigate') {
		event.respondWith(networkFirstNavigation(request));
		return;
	}

	if (path.includes('/__data.json')) {
		event.respondWith(networkFirstData(request));
		return;
	}

	if (path.startsWith('/api/')) {
		event.respondWith(networkFirstApi(request));
		return;
	}

	event.respondWith(cacheFirst(request));
});

async function cacheFirst(request: Request): Promise<Response> {
	const cached = await caches.match(request);
	if (cached) return cached;
	try {
		const response = await fetch(request);
		if (response.ok) {
			const cache = await caches.open(STATIC_CACHE);
			await cache.put(request, response.clone());
		}
		return response;
	} catch {
		return Response.error();
	}
}

async function networkFirstNavigation(request: Request): Promise<Response> {
	const cache = await caches.open(PAGES_CACHE);
	try {
		// `redirect: 'manual'` empêche le SW de suivre une redirection serveur
		// (ex: `/` → `/login` quand déconnecté) et de la servir en 200 sous l'URL
		// d'origine. On renvoie la redirection telle quelle pour que le navigateur
		// la suive et mette à jour la barre d'URL (sinon le formulaire de login
		// posterait sur `/` → 405 "No form actions exist for this page").
		const response = await fetch(request, { redirect: 'manual' });
		if (response.type === 'opaqueredirect' || (response.status >= 300 && response.status < 400)) {
			return response;
		}
		if (response.ok) {
			await cachePut(PAGES_CACHE, request, response.clone(), 20);
		}
		return response;
	} catch {
		const cached = await cache.match(request);
		if (cached) return cached;
		// Rediriger vers /offline pour que SvelteKit puisse hydrater correctement
		return Response.redirect(new URL('/offline', self.location.origin).href, 302);
	}
}

async function networkFirstData(request: Request): Promise<Response> {
	const cache = await caches.open(PAGES_CACHE);
	try {
		const response = await fetch(request);
		if (response.ok) {
			await cachePut(PAGES_CACHE, request, response.clone(), 20);
		}
		return response;
	} catch {
		const cached = await cache.match(request, { ignoreSearch: true });
		if (cached) return cached;
		// Retourner un redirect SvelteKit natif vers /offline
		// SvelteKit comprend ce format et fait une navigation client-side
		return new Response(
			JSON.stringify({ type: 'redirect', location: '/offline', status: 307 }),
			{ status: 200, headers: { 'Content-Type': 'application/json' } }
		);
	}
}

async function networkFirstApi(request: Request): Promise<Response> {
	const cache = await caches.open(API_CACHE);
	try {
		const response = await fetch(request);
		if (response.ok) {
			await cachePut(API_CACHE, request, response.clone(), 50, 7 * 24 * 60 * 60);
		}
		return response;
	} catch {
		const cached = await cache.match(request);
		return (
			cached ??
			new Response('{"error":"Hors ligne"}', {
				status: 503,
				headers: { 'Content-Type': 'application/json' }
			})
		);
	}
}
