<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';
	import { isSignedIn, posts, postsLoaded } from '../../../stores';

	let email = '';
	let password = '';
	let errorMessage = '';

	const signin = async () => {
		const response = await fetch(PUBLIC_BACKEND_API_URL + 'signin', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ email, password }),
			credentials: 'include'
		});

		if (response.ok) {
			// browser???
			if (browser) {
				posts.set(await response.json());
				postsLoaded.set(true);
			}
			isSignedIn.set(true);
			goto('/saved');
		} else {
			errorMessage = await response.text();
		}
	};
</script>

<form class="border-b-2 border-black border-dashed" on:submit|preventDefault={signin}>
	<input
		id="email"
		bind:value={email}
		required
		placeholder="Email"
		class="w-full py-1 mb-2 px-2 border-2 border-black"
	/>

	<input
		id="password"
		bind:value={password}
		type="password"
		required
		placeholder="Password"
		class="w-full py-1 px-2 border-2 border-black"
	/>

	<div class="flex items-center">
		<input
			type="submit"
			value="Sign in"
			class="py-1 px-2 my-4 bg-black text-white border-2 border-black hover:bg-gray-700 cursor-pointer"
		/>
		<span class="text-black ml-2">{errorMessage}</span>
	</div>
</form>

<p class="mt-4">
	Don't have an account? <a href="/register" class="text-black underline hover:text-gray-500"
		>Register.</a
	>
</p>

<svelte:head>
	<title>Sign In - Lucentsave</title>
</svelte:head>
