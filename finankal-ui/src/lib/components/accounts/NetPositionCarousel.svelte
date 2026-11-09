<script lang="ts">
    import PositionCard from "./PositionCard.svelte";
    import type { PositionTile, Slice } from "$lib/utils/payables";

    let { slides }: { slides: { tile: PositionTile; slices?: Record<number, Slice[]> }[] } = $props();

    let track: HTMLDivElement;
    let active = $state(0);

    // Wide screens show all three side by side. Narrower ones swipe, with the next card peeking in;
    // the dots there show where you are and jump to a card.
    function cardStep() {
        const first = track.children[0] as HTMLElement | undefined;
        const gap = parseFloat(getComputedStyle(track).columnGap) || 0;
        return first ? first.offsetWidth + gap : track.clientWidth;
    }

    function onScroll() {
        active = Math.min(slides.length - 1, Math.round(track.scrollLeft / cardStep()));
    }

    function go(index: number) {
        const smooth = !window.matchMedia("(prefers-reduced-motion: reduce)").matches;
        track.scrollTo({ left: index * cardStep(), behavior: smooth ? "smooth" : "auto" });
    }
</script>

<section aria-labelledby="where-you-stand">
    <h2 id="where-you-stand" class="sr-only">Where you stand</h2>

    <div
            bind:this={track}
            onscroll={onScroll}
            class="-mx-4 flex snap-x snap-mandatory gap-4 overflow-x-auto scroll-px-4 px-4 pb-1 [scrollbar-width:none] lg:mx-0 lg:grid lg:grid-cols-3 lg:overflow-visible lg:px-0 [&::-webkit-scrollbar]:hidden"
    >
        {#each slides as slide (slide.tile.question)}
            <div class="w-[85%] shrink-0 snap-start sm:w-[60%] lg:w-auto">
                <PositionCard tile={slide.tile} slices={slide.slices} />
            </div>
        {/each}
    </div>

    <div class="mt-4 flex justify-center gap-2 lg:hidden">
        {#each slides as slide, i (slide.tile.question)}
            <button
                    type="button"
                    onclick={() => go(i)}
                    aria-label={`Show "${slide.tile.question}"`}
                    aria-current={i === active}
                    class={`h-2 rounded-full transition-all focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 motion-reduce:transition-none ${i === active ? "w-6 bg-slate-300" : "w-2 bg-slate-600 hover:bg-slate-500"}`}
            ></button>
        {/each}
    </div>
</section>
