<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
    import Navbar from "$lib/components/layout/Navbar.svelte";

	let { children } = $props();
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<Navbar />

<main class="min-h-screen bg-slate-950 ">

    <div class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
        {@render children()}
    </div>

</main>