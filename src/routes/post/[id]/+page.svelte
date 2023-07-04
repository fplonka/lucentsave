<script lang="ts">
	import type { PageData } from './$types';
	import { markAsRead, like } from '$lib/postActions';
	import { goto } from '$app/navigation';

	export let data: PageData;

	let deleteState = 'Delete';
	let deleteTimeout: number;

	const initiateDelete = (postID: number) => {
		if (deleteState === 'Delete') {
			deleteState = 'Sure?';
			deleteTimeout = setTimeout(() => {
				deleteState = 'Delete';
			}, 3000); // Revert back to 'Delete' after 3 seconds
		} else if (deleteState === 'Sure?') {
			clearTimeout(deleteTimeout);
			deleteState = 'Delete';
			deletePost(postID);
		}
	};

	const reset = () => {
		deleteState = 'Delete';
	};

	const deletePost = async (postID: number): Promise<void> => {
		const response = await await fetch(`http://localhost:8080/api/deletePost?id=${postID}`, {
			method: 'DELETE',
			credentials: 'include'
		});
		if (response.ok) {
			goto('/saved');
		}
	};
</script>

<div class="space-y-4 mt-4">
	<div class="border-b-2 border-dashed border-black">
		<div on:mouseleave={reset} class="flex justify-between items-center group">
			<div>
				<h2 class="text-2xl font-bold text-black">{data.post.title}</h2>
				<a href={data.post.url} class="text-sm text-black block hover:underline hover:text-gray-500"
					>{data.post.url}</a
				>
			</div>
			<!-- <div
				on:click={initiateDelete}
				class="text-black text-xl p-2 font-bold cursor-pointer hover:text-gray-500 opacity-0 group-hover:opacity-100"
			>
				{deleteState}
			</div> -->
			<button
				class="opacity-0 group-hover:opacity-100 px-2 py-1 bg-black text-white border-2 border-black hover:bg-gray-700 cursor-pointer"
				on:click={() => initiateDelete(data.post.id)}
			>
				{deleteState}
			</button>
		</div>
		<div
			class="text-black mt-2 pb-4 prose prose-quoteless prose-blockquote:font-normal hover:prose-a:text-gray-500 relative"
		>
			{@html data.post.body}
		</div>
	</div>
</div>

<div class="space-y-4 mt-4">
	<div class="flex space-x-2">
		<button
			class="py-1 px-2 bg-black text-white border-2 border-black hover:bg-gray-700 cursor-pointer"
			on:click={async () => (data.post = await markAsRead(data.post))}
		>
			{data.post.isRead ? 'Mark as unread' : 'Mark as read'}
		</button>
		<span
			on:click={async () => (data.post = await like(data.post))}
			class="text-black px-2 py-1 cursor-pointer text-xl hover:text-gray-500"
			style="visibility: {data.post.isRead ? 'visible' : 'hidden'};"
		>
			{data.post.isLiked ? '★' : '☆'}
		</span>
	</div>
</div>
