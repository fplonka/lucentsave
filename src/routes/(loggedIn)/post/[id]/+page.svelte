<script lang="ts">
	import { goto } from '$app/navigation';
	import { PUBLIC_BACKEND_API_URL } from '$env/static/public';
	import { like, markAsRead } from '$lib/postActions';
	import { onMount } from 'svelte';
	import { posts } from '../../../../stores';
	import type { PageData } from './$types';
	import { v4 as uuid } from 'uuid';
	import { highlightRange, isNodeInRange } from '$lib/highlighting';
	import HighlightButton from './HighlightButton.svelte';
	import PostBody from '../../PostBody.svelte';

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

	let highlightButtonVisible: boolean = false;
	let highlightButtonPosition = { x: 0, y: 0 };

	let highlightDeleteButtonVisible: boolean = false;
	let highlightDeleteButtonPosition = { x: 0, y: 0 };

	let selectedHighlightId = '';

	const deleteHighlight = async (highlightId: string) => {
		// Remove all spans with this highlight ID
		highlightDeleteButtonVisible = false;
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

		reloadEventListeners();
	};

	onMount(() => {
		const mouseupHandler = async (event: MouseEvent) => {
			const userSelection = window.getSelection();
			// Check if the selection is within the "postbody" div
			const postBody = document.getElementById('postbody');
			if (
				userSelection &&
				userSelection.rangeCount > 0 &&
				userSelection.toString().length > 0 &&
				postBody?.contains(userSelection.anchorNode) &&
				postBody.contains(userSelection.focusNode) &&
				userSelection.toString().length > 0
			) {
				highlightButtonVisible = true;

				const bounds = document.getElementById('content')!.getBoundingClientRect();
				highlightButtonPosition = {
					x: event.clientX - bounds?.left,
					y: event.clientY - bounds?.top
				};
			}
		};

		const selectionchangeHandler = async (event: Event) => {
			const userSelection = window.getSelection();
			// Check if the selection is within the "postbody" div
			const postBody = document.getElementById('postbody');
			if (
				userSelection &&
				userSelection.rangeCount > 0 &&
				userSelection.toString().length > 0 &&
				postBody?.contains(userSelection.anchorNode) &&
				postBody.contains(userSelection.focusNode) &&
				userSelection.toString().length > 0
			) {
				const rect = userSelection.getRangeAt(0).getBoundingClientRect();
				const bounds = document.getElementById('content')!.getBoundingClientRect();

				highlightButtonVisible = true;
				const rootFontSize = parseFloat(getComputedStyle(document.documentElement).fontSize);
				highlightButtonPosition = {
					x: rect.left + rect.width / 2 - bounds.left,
					y: rect.bottom + 4 * rootFontSize - bounds.top // Offset the 3.5 rem from the button and still move
					// it down a bit. Super cursed
				};
			}
		};

		let clickHandler = async (event: MouseEvent) => {
			const target = event.target as HTMLElement;
			if (target.parentNode && target.parentNode.nodeName === 'A') {
				// User clicked on a link, do not delete the highlight.
				return;
			}

			if (target.dataset.highlightId) {
				const highlightId = target.dataset.highlightId;

				selectedHighlightId = highlightId;
				highlightDeleteButtonVisible = true;
				highlightSelected();
				const bounds = document.getElementById('content')!.getBoundingClientRect();
				highlightDeleteButtonPosition = {
					x: event.clientX - bounds?.left,
					y: event.clientY - bounds?.top
				};
			}
		};

		const mousedownHandler = (event: Event) => {
			const target = event.target as HTMLElement;
			let button = document.getElementById('highlightButton');

			if (highlightButtonVisible && (!button || !button.contains(target))) {
				highlightButtonVisible = false;
				if (window.getSelection()) {
					window.getSelection()!.empty();
				}
			}

			button = document.getElementById('highlightDeleteButton');
			if (highlightDeleteButtonVisible && (!button || !button.contains(target))) {
				unhighlightSelected();
				highlightDeleteButtonVisible = false;
			}
		};
		const isTouchDevice = 'ontouchstart' in window || navigator.maxTouchPoints > 0;

		// document.getElementById('postbody')!.addEventListener('mouseup', mouseupHandler);
		document.addEventListener('mousedown', mousedownHandler);
		document.addEventListener('click', clickHandler);

		if (isTouchDevice) {
			document.addEventListener('selectionchange', selectionchangeHandler);
		} else {
			document.addEventListener('mouseup', mouseupHandler);
		}

		reloadEventListeners();

		return () => {
			// document.getElementById('postbody')!.removeEventListener('mouseup', mouseupHandler);
			document.removeEventListener('mousedown', mousedownHandler);
			document.removeEventListener('click', clickHandler);

			if (isTouchDevice) {
				document.removeEventListener('selectionchange', selectionchangeHandler);
			} else {
				document.removeEventListener('mouseup', mouseupHandler);
			}
		};
	});

	let addHighlight = async () => {
		const userSelection = window.getSelection();
		highlightButtonVisible = false;
		// Check if the selection is within the "postbody" div
		if (userSelection && userSelection.rangeCount > 0 && userSelection.toString().length > 0) {
			let highlightID = uuid();
			highlightRange(userSelection.getRangeAt(0), highlightID);
			document.getSelection()?.empty();

			// Awful code to get the HTML of the smallest set of paragraphs which contanis the entire highlight.
			let highlights = Array.from(
				document.querySelectorAll(`span[data-highlight-id="${highlightID}"]`)
			);
			let paragraphs = new Set();
			highlights.forEach((highlight) => {
				// Find the closest paragraph parent of each highlight
				let parent = highlight.closest('p');
				if (parent) {
					paragraphs.add(parent.outerHTML);
				}
			});
			let paragraphHTML = Array.from(paragraphs).join('');

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
					text: paragraphHTML
				})
			});

			// Save highlighted post body
			await updateBody();

			reloadEventListeners();
		}
	};

	function reloadEventListeners() {
		// Select all highlighted spans
		let highlights = Array.from(document.querySelectorAll(`span[data-highlight-id]`));

		// Remove any previous listeners (to avoid duplication if this function is called more than once)
		highlights.forEach((highlight) => {
			highlight.removeEventListener('mouseenter', handleMouseEnter);
			highlight.removeEventListener('mouseleave', handleMouseLeave);
		});

		// Attach new event listeners
		highlights.forEach((highlight) => {
			highlight.addEventListener('mouseenter', handleMouseEnter);
			highlight.addEventListener('mouseleave', handleMouseLeave);
		});
	}

	// Event handler for mouseenter
	function handleMouseEnter(this: HTMLElement) {
		if (highlightDeleteButtonVisible) {
			return;
		}
		let highlightID = this.dataset.highlightId;
		let highlights = document.querySelectorAll(`span[data-highlight-id="${highlightID}"]`);
		highlights.forEach((highlight) => highlight.classList.add('bg-yellow-300'));
	}

	// Event handler for mouseleave
	function handleMouseLeave(this: HTMLElement) {
		if (highlightDeleteButtonVisible) {
			return;
		}
		let highlightID = this.dataset.highlightId;
		let highlights = document.querySelectorAll(`span[data-highlight-id="${highlightID}"]`);
		highlights.forEach((highlight) => highlight.classList.remove('bg-yellow-300'));
	}

	function highlightSelected() {
		let highlights = document.querySelectorAll(`span[data-highlight-id="${selectedHighlightId}"]`);
		highlights.forEach((highlight) => highlight.classList.add('bg-yellow-300'));
	}

	function unhighlightSelected() {
		let highlights = document.querySelectorAll(`span[data-highlight-id="${selectedHighlightId}"]`);
		highlights.forEach((highlight) => highlight.classList.remove('bg-yellow-300'));
	}
</script>

<HighlightButton
	bind:visible={highlightButtonVisible}
	position={highlightButtonPosition}
	callback={addHighlight}
	buttonText="＋"
	id="highlightButton"
/>

<HighlightButton
	bind:visible={highlightDeleteButtonVisible}
	position={highlightDeleteButtonPosition}
	callback={() => deleteHighlight(selectedHighlightId)}
	buttonText="✕"
	id="highlightDeleteButton"
/>

<div id="content" class="space-y-4 mt-4">
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
					if (confirm('Are you sure you want to delete this post? All highlights will be lost.')) {
						await deletePost(post.id);
					}
				}}
				class=" text-black px-2 py-1 cursor-pointer font-black hover:text-gray-500"
			>
				✕
			</span>
		</div>
		<PostBody classes="">
			{@html post.body}
		</PostBody>
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
