<script lang="ts">
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isSignedIn, posts } from '../stores';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';

	let dropdownOpen = false;

	const openDropdown = () => {
		dropdownOpen = true;
	};

	const closeDropdown = () => {
		dropdownOpen = false;
	};

	const signout = async () => {
		const response = await fetch(PUBLIC_BACKEND_API_URL + '/api/signout', {
			method: 'POST',
			credentials: 'include'
		});

		if (response.ok) {
			dropdownOpen = false;
			await goto('/signin');
			isSignedIn.set(false);
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
		console.log('getting posts into store');

		const fetchPosts = async () => {
			const response = await fetch(PUBLIC_BACKEND_API_URL + '/api/getAllUserPosts', {
				credentials: 'include'
			});
			if (response.ok) {
				posts.set(await response.json());
			}
		};
		fetchPosts();
	});
</script>

{#if $isSignedIn}
	<nav
		class="p-2 md:p-3 border-b-2 border-black"
		on:mouseleave={closeDropdown}
		on:click={closeDropdown}
	>
		<div class="container max-w-3xl mx-auto flex justify-between items-center">
			<div>
				<a href="/" class="font-extrabold text-l md:text-xl hover:text-gray-500">lucentsave</a>
			</div>
			<div class="flex justify-between items-center">
				<div role="navigation">
					<a
						href="/saved"
						class="hover:text-gray-500 text-black mr-2 md:mr-4 cursor-pointer {$page.url
							.pathname === '/saved'
							? 'font-bold'
							: ''}">Saved</a
					>
					<a
						href="/read"
						class="hover:text-gray-500 text-black mr-2 md:mr-4 cursor-pointer {$page.url
							.pathname === '/read'
							? 'font-bold'
							: ''}">Read</a
					>
					<a
						href="/liked"
						class="hover:text-gray-500 text-black mr-2 md:mr-4 cursor-pointer {$page.url
							.pathname === '/liked'
							? 'font-bold'
							: ''}">Liked</a
					>
				</div>
				<div role="menu" tabindex="0" class="relative" on:mouseenter={openDropdown}>
					<button class="relative z-10 mx-1 cursor-pointer">☰</button>
					{#if dropdownOpen}
						<div class="absolute right-0 w-36 bg-white z-20 border-2 border-black">
							<ul class="text-black shadow-box">
								<a
									href="/search"
									class="block cursor-pointer py-2 px-4 hover:text-gray-500 {$page.url.pathname ==
									'/search'
										? 'font-bold'
										: ''}">Search</a
								>
								<button class="cursor-pointer py-2 px-4 hover:text-gray-500" on:click={signout}>
									Sign out
								</button>
							</ul>
						</div>
					{/if}
				</div>
			</div>
		</div>
	</nav>
{/if}

<div class="mb-8">
	<div class="px-2 md:px-4 mx-auto max-w-2xl relative">
		<slot />
	</div>
</div>
