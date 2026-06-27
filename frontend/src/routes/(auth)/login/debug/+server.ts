import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

// DEBUG TEMPORAIRE (à retirer) : reçoit les infos collectées côté client sur la
// page login et les écrit dans les logs du container frontend (`make logs`).
// Chemin sous /login -> hooks.server.ts fait un early-return (pas d'auth), et le
// service worker ignore les POST : le beacon passe direct.
export const POST: RequestHandler = async ({ request }) => {
	const data = await request.json().catch(() => ({}));
	console.log('[LOGIN-DEBUG]', JSON.stringify(data));
	return json({ ok: true });
};
