<script lang="ts">
	export let classes: string = '';
</script>

<div
	id="postbody"
	class="prose
			prose-base md:prose-lg text-black mt-2 pb-4 prose-pre:rounded-none prose-pre:bg-gray-100 prose-pre:text-black
			prose-img:mx-auto prose-img:mb-1 prose-quoteless prose-blockquote:font-normal hover:prose-a:text-gray-500
			relative prose-code:before:hidden prose-code:after:hidden prose-code:bg-gray-100 prose-code:font-normal prose-code:p-0.5
		    {classes}"
>
	<slot />
</div>
