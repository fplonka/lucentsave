<script lang="ts">
	import { goto } from '$app/navigation';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';
	import { isSignedIn, postsLoaded } from '../../../stores';

	let email = ''; // Must be a valid email.
	let password = '';
	let confirmedPassword = '';
	let errorMessage = '';
	let formSubmitted = false;

	$: {
		errorMessage = '';
		if (email && !/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/.test(email)) {
			errorMessage = 'Invalid email';
		} else if (password !== confirmedPassword) {
			errorMessage = 'Passwords do not match';
		}
	}

	const register = async () => {
		formSubmitted = true;
		if (!errorMessage) {
			const response = await fetch(PUBLIC_BACKEND_API_URL + 'createUser', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ email, password }),
				credentials: 'include'
			});

			if (response.ok) {
				isSignedIn.set(true);
				postsLoaded.set(true);
				goto('/saved');
			} else {
				errorMessage = await response.text();
			}
		}
	};
</script>

<div>
	<form class="border-b-2 border-black border-dashed" on:submit|preventDefault={register}>
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
			placeholder="Password"
			required
			class="w-full py-1 mb-2 px-2 border-2 border-black"
		/>

		<!-- <label for="confirmedPassword" class="text-black">Confirm password:</label> -->
		<input
			id="confirmedPassword"
			bind:value={confirmedPassword}
			type="password"
			placeholder="Confirm password"
			required
			class="w-full py-1 px-2 border-2 border-black"
		/>

		<div class="flex items-center">
			<input
				type="submit"
				value="Register"
				class="py-1 px-2 my-4 bg-black text-white border-2 border-black hover:bg-gray-700 cursor-pointer"
			/>
			{#if formSubmitted}
				<span class="text-black ml-2">{errorMessage}</span>
			{/if}
		</div>
	</form>

	<p class="mt-4">
		Already have an account? <a href="/signin" class="text-black underline hover:text-gray-500"
			>Sign in.</a
		>
	</p>
</div>

<svelte:head>
	<title>Register - Lucentsave</title>
</svelte:head>
