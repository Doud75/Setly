import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit(),
		SvelteKitPWA({
			strategies: 'injectManifest',
			srcDir: 'src',
			filename: 'service-worker.ts',
			registerType: 'autoUpdate',
			includeAssets: [
				'apple-touch-icon.png',
				'web-app-manifest-192x192.png',
				'web-app-manifest-192x192-maskable.png',
				'web-app-manifest-512x512.png',
				'web-app-manifest-512x512-maskable.png'
			],
			manifest: {
				name: 'Setly',
				short_name: 'Setly',
				id: 'fr.setly.app',
				description: 'Une application pour gérer les setlists de votre groupe',
				start_url: '/',
				scope: '/',
				theme_color: '#f8fafc',
				background_color: '#000000',
				dir: 'ltr',
				display: 'standalone',
				display_override: ['window-controls-overlay', 'standalone'],
				orientation: 'portrait',
				prefer_related_applications: false,
				launch_handler: {
					client_mode: 'focus-existing'
				},
				edge_side_panel: {
					preferred_width: 400
				},
				categories: ['music', 'productivity'],
				shortcuts: [
					{
						name: 'Mes chansons',
						short_name: 'Chansons',
						description: 'Accéder à la bibliothèque de chansons',
						url: '/song',
					},
					{
						name: 'Nouvelle setlist',
						short_name: 'Setlist',
						description: 'Créer une nouvelle setlist',
						url: '/setlist/new',
					},
				],
				screenshots: [
					{ src: 'screenshots/mobile-dashboard.png', sizes: '390x844',  type: 'image/png', form_factor: 'narrow', label: 'Tableau de bord' },
					{ src: 'screenshots/mobile-songs.png',     sizes: '390x844',  type: 'image/png', form_factor: 'narrow', label: 'Liste des chansons' },
					{ src: 'screenshots/mobile-setlist.png',   sizes: '390x844',  type: 'image/png', form_factor: 'narrow', label: 'Détail setlist' },
					{ src: 'screenshots/mobile-song.png',      sizes: '390x844',  type: 'image/png', form_factor: 'narrow', label: 'Détail chanson' },
					{ src: 'screenshots/desktop-dashboard.png', sizes: '1280x800', type: 'image/png', form_factor: 'wide',   label: 'Tableau de bord' },
					{ src: 'screenshots/desktop-songs.png',     sizes: '1280x800', type: 'image/png', form_factor: 'wide',   label: 'Liste des chansons' },
					{ src: 'screenshots/desktop-setlist.png',   sizes: '1280x800', type: 'image/png', form_factor: 'wide',   label: 'Détail setlist' },
					{ src: 'screenshots/desktop-song.png',      sizes: '1280x800', type: 'image/png', form_factor: 'wide',   label: 'Détail chanson' },
				],
				icons: [
					{
						src: 'web-app-manifest-192x192.png',
						sizes: '192x192',
						type: 'image/png',
						purpose: 'any'
					},
					{
						src: 'web-app-manifest-192x192-maskable.png',
						sizes: '192x192',
						type: 'image/png',
						purpose: 'maskable'
					},
					{
						src: 'web-app-manifest-512x512.png',
						sizes: '512x512',
						type: 'image/png',
						purpose: 'any'
					},
					{
						src: 'web-app-manifest-512x512-maskable.png',
						sizes: '512x512',
						type: 'image/png',
						purpose: 'maskable'
					}
				]
			},
			devOptions: {
				enabled: true,
				type: 'module',
			},
		})
	]
});