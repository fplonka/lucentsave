<script>
	import '../app.css';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '../stores';
	import { onMount } from 'svelte';
	import { browser } from '$app/environment';

	let dropdownOpen = false;

	const openDropdown = () => {
		dropdownOpen = true;
	};

	const closeDropdown = () => {
		dropdownOpen = false;
	};

	const logout = async () => {
		const response = await fetch('http://localhost:8080/api/logout', {
			method: 'POST',
			credentials: 'include'
		});

		if (response.ok) {
			isLoggedIn.set(false); // If the logout was successful, update the isLoggedIn store
			dropdownOpen = false;
			goto('/login');
		} else {
			console.error('failed to log out');
			// TODO
		}
	};

	onMount(() => {
		if (browser) {
			const cookieExists = document.cookie
				.split(';')
				.some((item) => item.trim().startsWith('loggedIn='));
			isLoggedIn.set(cookieExists);
		}
	});
</script>

{#if $isLoggedIn}
	<nav class="p-2 md:p-3 border-b-2 border-black">
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
				<div class="relative" on:mouseenter={openDropdown} on:mouseleave={closeDropdown}>
					<button class="relative z-10 mx-1 cursor-pointer">☰</button>
					{#if dropdownOpen}
						<div class="absolute right-0 w-36 bg-white z-20 border-2 border-black">
							<ul class="text-black shadow-box">
								<li class="cursor-pointer py-2 px-4 hover:text-gray-500">Search</li>
								<li class="cursor-pointer py-2 px-4 hover:text-gray-500" on:click={logout}>
									Log out
								</li>
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

<style>
	details summary::-webkit-details-marker {
		display: none;
	}
	details summary:after {
		content: '☰';
	}
	html {
		overflow-y: scroll;
	}
</style>
