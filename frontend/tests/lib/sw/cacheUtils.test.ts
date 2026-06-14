import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { cachePut } from '$lib/sw/cacheUtils';

/**
 * Faux Cache (Cache API) backé par une Map, suffisant pour cachePut :
 * put / match / keys / delete, clés indexées par URL.
 */
class FakeCache {
	store = new Map<string, Response>();

	private keyOf(req: RequestInfo): string {
		return typeof req === 'string' ? req : req.url;
	}

	async put(req: RequestInfo, res: Response): Promise<void> {
		this.store.set(this.keyOf(req), res);
	}

	async match(req: RequestInfo): Promise<Response | undefined> {
		return this.store.get(this.keyOf(req));
	}

	async keys(): Promise<Request[]> {
		return [...this.store.keys()].map((url) => new Request(url));
	}

	async delete(req: RequestInfo): Promise<boolean> {
		return this.store.delete(this.keyOf(req));
	}
}

const cacheInstances = new Map<string, FakeCache>();

const fakeCaches = {
	open: async (name: string): Promise<FakeCache> => {
		if (!cacheInstances.has(name)) cacheInstances.set(name, new FakeCache());
		return cacheInstances.get(name)!;
	}
};

function cacheFor(name: string): FakeCache {
	return cacheInstances.get(name)!;
}

const BASE = new Date('2026-06-14T12:00:00.000Z').getTime();

beforeEach(() => {
	cacheInstances.clear();
	vi.stubGlobal('caches', fakeCaches);
	vi.useFakeTimers();
	vi.setSystemTime(BASE);
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
});

describe('cachePut', () => {
	it('ajoute un header sw-cached-at avec le timestamp courant', async () => {
		await cachePut('c', 'https://x.test/1', new Response('a'), 10);

		const cached = await cacheFor('c').match('https://x.test/1');
		expect(cached?.headers.get('sw-cached-at')).toBe(String(BASE));
	});

	it('conserve le statut de la réponse mise en cache', async () => {
		await cachePut('c', 'https://x.test/1', new Response('a', { status: 201 }), 10);

		const cached = await cacheFor('c').match('https://x.test/1');
		expect(cached?.status).toBe(201);
	});

	it('plafonne le cache à maxEntries en évinçant les plus anciennes (LRU)', async () => {
		for (let i = 0; i < 5; i++) {
			vi.setSystemTime(BASE + i * 1000);
			await cachePut('c', `https://x.test/${i}`, new Response('a'), 3);
		}

		const urls = [...cacheFor('c').store.keys()].sort();
		expect(urls).toHaveLength(3);
		// Les 2 plus anciennes (0 et 1) sont évincées, les 3 récentes conservées.
		expect(urls).toEqual([
			'https://x.test/2',
			'https://x.test/3',
			'https://x.test/4'
		]);
	});

	it('ne supprime rien tant que le nombre d’entrées est sous maxEntries', async () => {
		for (let i = 0; i < 3; i++) {
			vi.setSystemTime(BASE + i * 1000);
			await cachePut('c', `https://x.test/${i}`, new Response('a'), 5);
		}

		expect(cacheFor('c').store.size).toBe(3);
	});

	it('évince les entrées expirées selon le TTL (maxAgeSeconds)', async () => {
		const sevenDays = 7 * 24 * 60 * 60;

		// Entrée ancienne
		vi.setSystemTime(BASE);
		await cachePut('c', 'https://x.test/old', new Response('a'), 50, sevenDays);

		// 8 jours plus tard, nouvelle entrée → l’ancienne doit expirer
		vi.setSystemTime(BASE + 8 * 24 * 60 * 60 * 1000);
		await cachePut('c', 'https://x.test/new', new Response('b'), 50, sevenDays);

		const urls = [...cacheFor('c').store.keys()];
		expect(urls).toEqual(['https://x.test/new']);
	});

	it('conserve les entrées encore valides au regard du TTL', async () => {
		const sevenDays = 7 * 24 * 60 * 60;

		vi.setSystemTime(BASE);
		await cachePut('c', 'https://x.test/a', new Response('a'), 50, sevenDays);

		// 1 jour plus tard : toujours valide
		vi.setSystemTime(BASE + 24 * 60 * 60 * 1000);
		await cachePut('c', 'https://x.test/b', new Response('b'), 50, sevenDays);

		const urls = [...cacheFor('c').store.keys()].sort();
		expect(urls).toEqual(['https://x.test/a', 'https://x.test/b']);
	});
});
