<script lang="ts">
	import '../app.css';
	import { onMount } from 'svelte';
	import favicon from '$lib/assets/favicon.svg';
	import PWABadge from '$lib/components/PWABadge.svelte';
	import OfflineIndicator from '$lib/components/OfflineIndicator.svelte';
	import { modalStore } from '$lib/stores/modalStore';
	import Modal from '$lib/components/ui/Modal.svelte';

	let { children } = $props();

	onMount(() => {
		// iOS Safari ne synthétise un `click` au tap que sur les éléments qu'il juge
		// « clickable ». Couplé à la délégation d'événements de Svelte 5, certains
		// boutons (ex: « Se connecter ») ne répondent qu'au 2e tap, ou pas du tout,
		// sur iPhone — alors qu'à la souris tout marche. Attacher un listener tactile
		// vide au document force iOS à émettre les `click` sur tous les éléments.
		// cf. sveltejs/svelte#13339
		const noop = () => {};
		document.addEventListener('touchstart', noop, { passive: true });
		return () => document.removeEventListener('touchstart', noop);
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<PWABadge/>
<OfflineIndicator />

{@render children?.()}

{#if $modalStore.isOpen && $modalStore.component}
	<Modal isOpen={$modalStore.isOpen} onClose={modalStore.close}>
		{@const Component = $modalStore.component}
		<Component {...$modalStore.props} close={modalStore.close} />
	</Modal>
{/if}