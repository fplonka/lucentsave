<script lang="ts">
	import { goto } from '$app/navigation';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';
	import { isSignedIn, postsLoaded } from '../../stores';

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

<div class="mt-12 mb-12 mx-auto w-max border-black border-2 p-8">
	<div class="space-y-2">
		<div class="flex items-baseline font-semibold">
			<span class="text-black text-2xl font-extrabold mr-2">Lucent</span>
			<span class="text-black text-base font-light">[ˈlü-sᵊnt]</span>
			<span class="text-black text-base font-semibold italic ml-2">adjective</span>
		</div>
		<div class="flex">
			<span class="font-bold mr-1 w-4 inline-block">1.</span>
			<span class=" text-black">Glowing with light</span>
		</div>
		<div class="flex">
			<span class="font-bold mr-1 w-4 inline-block">2.</span>
			<span class=" text-black">Marked by clarity or translucence</span>
		</div>
	</div>
</div>
<div class="font-normal text-black mb-12 mx-auto text-center text-base">
	A dead simple website for saving the things you want to read.
</div>

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
				class="py-1 px-2 my-4 bg-black text-white border-2 border-black hover:bg-gray-500 cursor-pointer"
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
