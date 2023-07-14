<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isSignedIn, posts, postsLoaded } from '../stores';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';

	const signout = async () => {
		const response = await fetch(PUBLIC_BACKEND_API_URL + 'signout', {
			method: 'POST',
			credentials: 'include'
		});

		if (response.ok) {
			dropdownOpen = false;
			await goto('/signin');
			postsLoaded.set(false);
			posts.set([]);
		} else {
			// TODO
		}
	};

	let dropdownOpen = false;
</script>

<nav class="p-3 border-b-2 border-black">
	<div class="container max-w-3xl xl:max-w-4xl mx-auto flex items-center justify-between">
		<a href="/" class="font-black text-lg sm:text-xl md:text-2xl hover:text-gray-500">Lucentsave</a>
		<div class="flex items-center">
			<a
				href="/saved"
				class="hover:text-gray-500 lg:text-lg text-black mr-2 sm:mr-4 cursor-pointer {$page.url
					.pathname === '/saved'
					? 'font-bold'
					: ''}">Saved</a
			>
			<a
				href="/read"
				class="hover:text-gray-500 lg:text-lg text-black mr-2 sm:mr-4 cursor-pointer {$page.url
					.pathname === '/read'
					? 'font-bold'
					: ''}">Read</a
			>
			<a
				href="/liked"
				class="hover:text-gray-500 lg:text-lg text-black mr-2 md:sm-4 cursor-pointer {$page.url
					.pathname === '/liked'
					? 'font-bold'
					: ''}">Liked</a
			>
			<div class="relative">
				<button
					on:click|stopPropagation={() => (dropdownOpen = !dropdownOpen)}
					class="text-lg lg:text-xl cursor-pointer z-10 px-1 sm:mx-1 hover:text-gray-500">☰</button
				>
				{#if dropdownOpen}
					<div class="absolute right-0 text-right w-max bg-white z-20 border-2 border-black mt-1">
						<ul class="text-black shadow-box">
							<a
								href="/search"
								class="lg:text-lg block cursor-pointer py-2 px-4 hover:text-gray-500 {$page.url
									.pathname === '/search'
									? 'font-bold'
									: ''}">Search</a
							>
							<a
								href="/highlights"
								class="lg:text-lg block cursor-pointer py-2 px-4 hover:text-gray-500 {$page.url
									.pathname === '/highlights'
									? 'font-bold'
									: ''}">Highlights</a
							>
							<button
								class="lg:text-lg w-full text-right block cursor-pointer py-2 px-4 hover:text-gray-500"
								on:click={signout}>Sign out</button
							>
						</ul>
					</div>
				{/if}
			</div>
		</div>
	</div>
</nav>

<svelte:window
	on:click={() => {
		if (dropdownOpen) {
			dropdownOpen = false;
		}
	}}
/>

<slot />
