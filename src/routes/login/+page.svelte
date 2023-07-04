<script lang="ts">
	import { goto } from '$app/navigation';
	import { isLoggedIn } from '../../stores';

	let email = '';
	let password = '';

	const login = async () => {
		const response = await fetch('http://localhost:8080/api/loginUser', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ username: email, password }),
			credentials: 'include'
		});

		if (response.ok) {
			isLoggedIn.set(true);
			goto('/saved');
		} else {
			alert('Login failed!');
		}
	};
</script>

<div class="mt-12 mb-10 mx-auto w-max border-black border-2 p-8">
	<div class="space-y-2">
		<div class="flex items-baseline font-semibold">
			<span class="text-black text-2xl font-extrabold mr-2">Lucent</span>
			<span class="text-black text-base font-light">[ˈlü-sᵊnt]</span>
			<span class="text-black text-base font-semibold italic ml-2">adjective</span>
		</div>
		<div class="flex">
			<span class="font-bold mr-1 w-4 inline-block">1.</span>
			<span class=" text-black">Glowing with light : Luminous</span>
		</div>
		<div class="flex">
			<span class="font-bold mr-1 w-4 inline-block">2.</span>
			<span class=" text-black">Marked by clarity or translucence : Clear</span>
		</div>
	</div>
</div>
<div class="font-normal text-black mb-8 mx-auto text-center text-base">
	A dead simple website for saving the things you want to read.
</div>

<div>
	<form class="border-b-2 border-black border-dashed" on:submit|preventDefault={login}>
		<h2 class="text-2xl mb-2 font-bold text-black">Login</h2>

		<label for="email" class="text-black">Email:</label>
		<input
			id="email"
			bind:value={email}
			required
			class="w-full py-1 mb-2 px-2 border-2 border-black"
		/>

		<label for="password" class="text-black">Password:</label>
		<input
			id="password"
			bind:value={password}
			type="password"
			required
			class="w-full py-1 px-2 border-2 border-black"
		/>

		<input
			type="submit"
			value="Login"
			class="py-1 px-2 my-4 bg-black text-white border-2 border-black hover:bg-gray-700 cursor-pointer"
		/>
	</form>

	<p class="mt-4">
		Don't have an account? <a href="/register" class="text-black underline hover:text-gray-500"
			>Register</a
		>
	</p>
</div>
