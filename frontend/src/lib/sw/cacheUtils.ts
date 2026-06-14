/**
 * Stocke une réponse dans un cache de la Cache API avec une stratégie d'éviction
 * LRU (Least Recently Used, basée sur l'horodatage d'écriture) et un TTL optionnel.
 *
 * - Ajoute un header `sw-cached-at` (timestamp ms) à la réponse mise en cache.
 * - Si `maxAgeSeconds` est fourni, supprime les entrées expirées.
 * - Si le nombre d'entrées dépasse `maxEntries`, supprime les plus anciennes.
 */
export async function cachePut(
	cacheName: string,
	request: RequestInfo,
	response: Response,
	maxEntries: number,
	maxAgeSeconds?: number
): Promise<void> {
	const cache = await caches.open(cacheName);

	// Cloner la réponse en ajoutant un header sw-cached-at (timestamp ms).
	// On reconstruit la Response car les headers d'une response existante sont immuables.
	const headers = new Headers(response.headers);
	headers.set('sw-cached-at', Date.now().toString());
	const timedResponse = new Response(response.body, {
		status: response.status,
		statusText: response.statusText,
		headers
	});
	await cache.put(request, timedResponse);

	// Récupérer toutes les clés avec leur timestamp d'écriture.
	const keys = await cache.keys();
	const entries = await Promise.all(
		keys.map(async (key) => {
			const res = await cache.match(key);
			return { key, cachedAt: Number(res?.headers.get('sw-cached-at') ?? 0) };
		})
	);

	const now = Date.now();
	const deleted = new Set<Request>();

	// TTL : supprimer les entrées expirées.
	if (maxAgeSeconds !== undefined) {
		const maxAgeMs = maxAgeSeconds * 1000;
		for (const entry of entries) {
			if (entry.cachedAt && now - entry.cachedAt > maxAgeMs) {
				await cache.delete(entry.key);
				deleted.add(entry.key);
			}
		}
	}

	// LRU : supprimer les plus anciennes au-delà de maxEntries.
	const remaining = entries
		.filter((e) => !deleted.has(e.key))
		.sort((a, b) => a.cachedAt - b.cachedAt);
	if (remaining.length > maxEntries) {
		const toDelete = remaining.slice(0, remaining.length - maxEntries);
		for (const entry of toDelete) {
			await cache.delete(entry.key);
		}
	}
}
