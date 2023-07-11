<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';
	import { like, markAsRead } from '$lib/postActions';
	import { onMount } from 'svelte';
	import { posts } from '../../../stores';
	import type { Post } from '../../[listType]/+page.server';
	import type { PageData } from './$types';
	import { v4 as uuid } from 'uuid';
	import { highlightRange, isNodeInRange } from '$lib/highlighting';

	export let data: PageData;

	let post = data.post;

	const deletePost = async (postID: number): Promise<void> => {
		const response = await await fetch(PUBLIC_BACKEND_API_URL + `deletePost?id=${postID}`, {
			method: 'DELETE',
			credentials: 'include'
		});
		if (response.ok) {
			const postIndex = $posts.findIndex((p) => p.id === post.id);
			posts.update((currentPosts) => currentPosts.filter((_, i) => i !== postIndex));
			goto('/saved');
		}
	};

	const updateBody = async () => {
		fetch(PUBLIC_BACKEND_API_URL + `updatePostBody`, {
			method: 'PUT',
			credentials: 'include',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ id: post.id, body: document.getElementById('postbody')?.innerHTML! })
		});
	};

	onMount(() => {
		const mouseupHandler = async () => {
			const userSelection = window.getSelection();
			// Check if the selection is within the "postbody" div
			const postBody = document.getElementById('postbody');
			if (
				userSelection &&
				userSelection.rangeCount > 0 &&
				userSelection.toString().length > 0 &&
				postBody?.contains(userSelection.anchorNode) &&
				postBody.contains(userSelection.focusNode)
			) {
				let highlightID = uuid();
				const selectedText = userSelection.toString();
				highlightRange(userSelection.getRangeAt(0), highlightID);
				document.getSelection()?.empty();

				// Store the created highlight on the backend
				await fetch(PUBLIC_BACKEND_API_URL + 'createHighlight', {
					method: 'PUT',
					credentials: 'include',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({
						id: highlightID,
						postId: parseInt(data.id),
						text: selectedText
					})
				});

				// Save highlighted post body
				await updateBody();
			}
		};

		const clickHandler = async (event: Event) => {
			const target = event.target as HTMLElement;
			if (target.dataset.highlightId) {
				const highlightId = target.dataset.highlightId;
				// Use highlightId to delete the highlight from backend

				// Remove all spans with this highlight ID
				const highlights = document.querySelectorAll(`span[data-highlight-id="${highlightId}"]`);
				highlights.forEach((span) => {
					const parent = span.parentNode;
					while (span.firstChild) {
						parent?.insertBefore(span.firstChild, span);
					}
					parent?.removeChild(span);
				});

				await fetch(PUBLIC_BACKEND_API_URL + `deleteHighlight?id=${highlightId}`, {
					method: 'PUT',
					credentials: 'include'
				});

				await updateBody();
			}
		};

		document.addEventListener('mouseup', mouseupHandler);
		document.addEventListener('click', clickHandler);

		return () => {
			document.removeEventListener('mouseup', mouseupHandler);
			document.removeEventListener('click', clickHandler);
		};
	});
</script>

<div class="space-y-4 mt-4">
	<div class="border-b-2 border-dashed border-black overflow-auto break-words">
		<div class="flex justify-between items-center group">
			<div>
				<h2 class="text-xl md:text-2xl font-bold text-black">{post.title}</h2>
				<a href={post.url} class="text-sm text-black block hover:underline hover:text-gray-500"
					>{post.url}</a
				>
			</div>
			<span
				on:click={async () => {
					if (confirm('Are you sure you want to delete this post?')) {
						await deletePost(post.id);
					}
				}}
				class=" text-black px-2 py-1 cursor-pointer font-black hover:text-gray-500"
			>
				✕
			</span>
		</div>
		<div
			id="postbody"
			class="prose
			prose-base md:prose-lg text-black mt-2 pb-4 prose-pre:rounded-none prose-pre:bg-gray-100 prose-pre:text-black
			prose-img:mx-auto prose-img:mb-1 prose-quoteless prose-blockquote:font-normal hover:prose-a:text-gray-500
			relative prose-code:before:hidden prose-code:after:hidden prose-code:bg-gray-100 prose-code:font-normal prose-code:p-0.5
			prose-
		
		"
		>
			{@html post.body}
		</div>
	</div>
</div>

<div class="space-y-4 mt-4">
	<div class="flex justify-between items-center">
		<div class="flex space-x-2">
			<button
				class="py-1 px-2 bg-black text-white border-2 border-black hover:bg-gray-700 cursor-pointer"
				on:click={async () => {
					// Doing this to make things more responsive: we update the client-side state instantly
					const postCopy = { ...post };
					post.isRead = !post.isRead;

					const postIndex = $posts.findIndex((p) => p.id === post.id);
					if (postIndex !== -1) {
						$posts[postIndex].isRead = !$posts[postIndex].isRead;
					}

					post = await markAsRead(postCopy);
					$posts[postIndex] = post;
				}}
			>
				{post.isRead ? 'Mark as Unread' : 'Mark as Read'}
			</button>
			<span
				on:click={async () => {
					const postCopy = { ...post };
					post.isLiked = !post.isLiked;

					const postIndex = $posts.findIndex((p) => p.id === post.id);
					if (postIndex !== -1) {
						$posts[postIndex].isLiked = !$posts[postIndex].isLiked;
					}

					post = await like(postCopy);
					$posts[postIndex] = post;
				}}
				class="text-black px-2 py-1 cursor-pointer text-xl hover:text-gray-500"
				style="visibility: {post.isRead ? 'visible' : 'hidden'};"
			>
				{post.isLiked ? '★' : '☆'}
			</span>
		</div>
		<span
			class="text-black px-2 py-1 cursor-pointer text-xl hover:text-gray-500"
			on:click={() => window.scrollTo(0, 0)}
		>
			↑
		</span>
	</div>
</div>

<svelte:head>
	<title>{post.title} - Lucentsave</title>
</svelte:head>
