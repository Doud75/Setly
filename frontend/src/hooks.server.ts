import type { Handle } from '@sveltejs/kit';
import { jwtDecode } from 'jwt-decode';
import { env } from '$env/dynamic/private';

type UserPayload = {
    user_id: number;
    exp: number;
};

const BACKEND_URL = env.BACKEND_INTERNAL_URL || 'http://backend:8089/api';

const AUTH_ROUTES = ['/login', '/signup'];
const PUBLIC_ROUTES = ['/offline'];

type RefreshResult = {
    token: string;
    refresh_token: string;
    bands?: { id: number }[];
};

// Single-flight du refresh : le frontend tourne en un seul process Node, donc une Map
// en mémoire suffit pour coalescer les refresh concurrents. Sans ça, les N requêtes
// /api/* parallèles déclenchaient chacune une rotation du refresh token côté backend,
// et les perdantes recevaient "refresh token not found" -> déconnexion.
const refreshInFlight = new Map<string, Promise<RefreshResult | null>>();
const refreshCache = new Map<string, { result: RefreshResult; expiresAt: number }>();
const REFRESH_CACHE_TTL = 60_000; // ~60s : couvre la fenêtre de propagation du cookie au client

// Appelle /auth/refresh une seule fois. Ne touche pas aux cookies (l'appelant s'en charge).
async function doRefresh(refreshToken: string, clientIp: string): Promise<RefreshResult | null> {
    try {
        const headers: Record<string, string> = { 'Content-Type': 'application/json' };
        if (clientIp) {
            headers['X-Forwarded-For'] = clientIp;
        }

        const refreshResponse = await fetch(`${BACKEND_URL}/auth/refresh`, {
            method: 'POST',
            headers,
            body: JSON.stringify({ refresh_token: refreshToken }),
            signal: AbortSignal.timeout(5000)
        });

        if (refreshResponse.ok) {
            return (await refreshResponse.json()) as RefreshResult;
        }

        console.error('[AUTH] Refresh failed with status:', refreshResponse.status);
        return null;
    } catch (error) {
        console.error('[AUTH] Refresh error:', error);
        return null;
    }
}

// Coalesce les refresh pour un même token : un seul appel réseau partagé par toutes
// les requêtes concurrentes, et un court cache pour celles arrivant juste après.
async function getRefreshResult(refreshToken: string, clientIp: string): Promise<RefreshResult | null> {
    // élagage paresseux des entrées expirées (la map reste minuscule)
    const now = Date.now();
    for (const [key, entry] of refreshCache) {
        if (entry.expiresAt <= now) {
            refreshCache.delete(key);
        }
    }

    const cached = refreshCache.get(refreshToken);
    if (cached && cached.expiresAt > now) {
        return cached.result;
    }

    const inflight = refreshInFlight.get(refreshToken);
    if (inflight) {
        return inflight;
    }

    const promise = doRefresh(refreshToken, clientIp)
        .then((result) => {
            // on ne cache pas les échecs : un vrai 401 reste réessayable
            if (result) {
                refreshCache.set(refreshToken, { result, expiresAt: Date.now() + REFRESH_CACHE_TTL });
            }
            return result;
        })
        .finally(() => refreshInFlight.delete(refreshToken));

    refreshInFlight.set(refreshToken, promise);
    return promise;
}

export const handle: Handle = async ({ event, resolve }) => {
    if (event.url.pathname === '/logout') {
        return resolve(event);
    }

    if (AUTH_ROUTES.some((r) => event.url.pathname.startsWith(r)) || PUBLIC_ROUTES.includes(event.url.pathname)) {
        return resolve(event);
    }

    const token = event.cookies.get('jwt_token');
    const refreshToken = event.cookies.get('refresh_token');
    const activeBandId = event.cookies.get('active_band_id');

    event.locals.token = token || null;
    event.locals.user = null;
    event.locals.activeBandId = activeBandId;

    let decoded: UserPayload | null = null;
    let needsRefresh = false;

    if (token) {
        try {
            decoded = jwtDecode<UserPayload>(token);
            const expiresIn = decoded.exp * 1000 - Date.now();
            needsRefresh = expiresIn < 0;
        } catch {
            needsRefresh = true;
        }
    }

    if ((!token || needsRefresh) && refreshToken) {
        let clientIp = '';
        try {
            clientIp = event.getClientAddress();
        } catch {
            /* ignore */
        }

        const refreshData = await getRefreshResult(refreshToken, clientIp);

        if (refreshData) {
            const cookieOptions = {
                path: '/',
                httpOnly: true,
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax' as const
            };

            event.cookies.set('jwt_token', refreshData.token, {
                ...cookieOptions,
                maxAge: 60 * 60 * 24 * 30
            });

            event.cookies.set('refresh_token', refreshData.refresh_token, {
                ...cookieOptions,
                maxAge: 60 * 60 * 24 * 30
            });

            if (refreshData.bands && refreshData.bands.length > 0) {
                event.cookies.set('user_bands', JSON.stringify(refreshData.bands), {
                    ...cookieOptions,
                    maxAge: 60 * 60 * 24 * 30
                });
                const currentBandId = activeBandId ? parseInt(activeBandId) : null;
                const stillValid = currentBandId && refreshData.bands.some((b: { id: number }) => b.id === currentBandId);
                const newActiveBandId = stillValid ? currentBandId!.toString() : refreshData.bands[0].id.toString();
                event.cookies.set('active_band_id', newActiveBandId, {
                    ...cookieOptions,
                    maxAge: 60 * 60 * 24 * 30
                });
            }

            event.locals.token = refreshData.token;
            decoded = jwtDecode<UserPayload>(refreshData.token);
        }
    }

    if (decoded && decoded.exp * 1000 > Date.now()) {
        try {
            const userInfoUrl = `${BACKEND_URL}/user/info`;
            const headers: Record<string, string> = {
                'Authorization': `Bearer ${event.locals.token || token}`
            };
            
            try {
                headers['X-Forwarded-For'] = event.getClientAddress();
            } catch { /* ignore */ }
            
            if (activeBandId) {
                headers['X-Band-ID'] = activeBandId;
            }
            const userInfoRes = await fetch(userInfoUrl, { headers, signal: AbortSignal.timeout(1500) });

            if (userInfoRes.ok) {
                const userInfo = await userInfoRes.json();
                event.locals.user = {
                    id: decoded.user_id,
                    username: userInfo.username,
                    band_name: userInfo.band_name,
                    role: userInfo.role
                };
            }
        } catch (error) {
            console.error('[AUTH] Failed to fetch user info:', error);
        }
    }

    return resolve(event);
};