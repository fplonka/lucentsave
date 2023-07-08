<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isSignedIn, posts, postsLoaded } from '../stores';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';

	let dropdownOpen = false;

	const signout = async () => {
		const response = await fetch(PUBLIC_BACKEND_API_URL + 'signout', {
			method: 'POST',
			credentials: 'include'
		});

		if (response.ok) {
			dropdownOpen = false;
			await goto('/signin');
			isSignedIn.set(false);
			postsLoaded.set(false);
			posts.set([]);
		} else {
			// TODO
		}
	};

	onMount(() => {
		if (browser) {
			const cookieExists = document.cookie
				.split(';')
				.some((item) => item.trim().startsWith('loggedIn='));
			isSignedIn.set(cookieExists);
		}
	});
</script>

{#if $isSignedIn}
	<nav class="p-3 border-b-2 border-black">
		<div class="container max-w-3xl xl:max-w-4xl mx-auto flex items-center justify-between">
			<a href="/" class="font-extrabold text-xl hover:text-gray-500">lucentsave</a>
			<div class="flex items-center">
				<a
					href="/saved"
					class="hover:text-gray-500 text-black mr-2 sm:mr-4 cursor-pointer {$page.url.pathname ===
					'/saved'
						? 'font-bold'
						: ''}">Saved</a
				>
				<a
					href="/read"
					class="hover:text-gray-500 text-black mr-2 sm:mr-4 cursor-pointer {$page.url.pathname ===
					'/read'
						? 'font-bold'
						: ''}">Read</a
				>
				<a
					href="/liked"
					class="hover:text-gray-500 text-black mr-2 md:sm-4 cursor-pointer {$page.url.pathname ===
					'/liked'
						? 'font-bold'
						: ''}">Liked</a
				>
				<div class="relative">
					<button
						on:click={() => (dropdownOpen = !dropdownOpen)}
						class="text-lg cursor-pointer z-10 px-1">☰</button
					>
					{#if dropdownOpen}
						<div class="absolute right-0 w-36 bg-white z-20 border-2 border-black mt-1">
							<ul class="text-black shadow-box">
								<a
									href="/search"
									class="block cursor-pointer py-2 px-4 hover:text-gray-500 {$page.url.pathname ===
									'/search'
										? 'font-bold'
										: ''}">Search</a
								>
								<button
									class="block cursor-pointer py-2 px-4 hover:text-gray-500"
									on:click={signout}>Sign out</button
								>
							</ul>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</nav>
{/if}

<div
	class="mb-8"
	on:click={() => {
		dropdownOpen = false;
	}}
>
	<div class="px-4 sm:px-6 mx-auto max-w-2xl xl:max-w-3xl relative">
		<slot />
	</div>
</div>
