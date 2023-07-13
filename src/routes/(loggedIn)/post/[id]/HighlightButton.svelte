<script lang="ts">
	export let visible = false;
	export let position = { x: 0, y: 0 };
	export let callback: ((event: Event) => Promise<void>) | (() => Promise<void>);
	export let buttonText: string = '';
	export let id: string = '';

	$: style = `position: absolute; left: ${position.x}px; top: ${position.y}px; z-index: 20;`;
</script>

<div>
	{#if visible}
		<button
			{id}
			{style}
			on:click|stopPropagation={callback}
			class="text-black -mt-14 bg-white border-black border-2 py-1 w-10 text-center text-xl font-black hover:text-gray-500"
		>
			{buttonText}
		</button>
	{/if}
</div>
