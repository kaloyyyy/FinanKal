<script lang="ts">
    import { goto } from "$app/navigation";
    import { currency } from "$lib/utils/currency";
    import { formatMonth } from "$lib/utils/date";
    import { moneyFlowGraph, type GraphNode } from "$lib/utils/moneyFlowGraph";

    let { data } = $props();

    const graph = $derived(moneyFlowGraph(data.accounts, data.flows));
    let focus = $state<string | null>(null); // hovered / focused account: dims everything it doesn't touch

    // Account type → colour; arrows take the colour of where the money came from.
    const COLORS: Record<string, string> = {
        INCOME: "#34d399",
        EQUITY: "#a3a3a3",
        ASSET: "#60a5fa",
        CREDIT_CARD: "#f472b6",
        EXPENSE: "#fb7185"
    };
    const color = (type: string) => COLORS[type] ?? "#94a3b8";

    const touches = (from: string, to: string) => !focus || focus === from || focus === to;
    const through = (n: GraphNode) => Math.max(n.inflow, n.outflow);

    function labelPosition(n: GraphNode) {
        if (n.column === 0) return { x: n.x - n.r - 10, y: n.y, anchor: "end" };
        if (n.column === 2) return { x: n.x + n.r + 10, y: n.y, anchor: "start" };
        return { x: n.x, y: n.y + n.r + 16, anchor: "middle" };
    }

    function pickMonth(event: Event) {
        const month = (event.currentTarget as HTMLSelectElement).value;
        goto(month ? `?month=${month}` : "?", { keepFocus: true, noScroll: true });
    }
</script>

<svelte:head>
    <title>Money flow</title>
</svelte:head>

<div class="space-y-8">
    <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight text-white">Money flow</h1>
            <p class="mt-1 text-slate-400">
                Where your money comes from and where it goes. Bigger circles move more money; thicker arrows carry more.
            </p>
        </div>
        <label class="flex items-center gap-3 text-sm text-slate-400">
            Period
            <select
                    value={data.month}
                    onchange={pickMonth}
                    class="rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-white outline-none focus:border-blue-500"
            >
                <option value="">All time</option>
                {#each data.months as month}
                    <option value={month}>{formatMonth(`${month}-01`)}</option>
                {/each}
            </select>
        </label>
    </header>

    {#if graph.edges.length}
        <section class="rounded-2xl border border-slate-800 bg-slate-900 p-4 sm:p-6">
            <div class="mb-2 grid grid-cols-3 text-xs font-semibold tracking-wider text-slate-500 uppercase">
                <span class="text-left">Comes from</span>
                <span class="text-center">Your accounts</span>
                <span class="text-right">Goes to</span>
            </div>

            <div class="overflow-x-auto">
                <svg
                        viewBox={`0 0 ${graph.width} ${graph.height}`}
                        class="h-auto w-full min-w-[640px]"
                        role="img"
                        aria-label="Money flow between your accounts"
                >
                    {#each graph.edges as edge (edge.id)}
                        <g
                                opacity={touches(edge.from.id, edge.to.id) ? 0.85 : 0.08}
                                class="transition-opacity"
                        >
                            <title>{edge.from.name} → {edge.to.name}: {currency.format(edge.amount)} ({edge.transactions} {edge.transactions === 1 ? "transaction" : "transactions"})</title>
                            <path
                                    d={edge.path}
                                    fill="none"
                                    stroke={color(edge.from.type)}
                                    stroke-width={edge.width}
                                    stroke-linecap="round"
                            />
                            <polygon points={edge.arrow} fill={color(edge.from.type)} />
                        </g>
                    {/each}

                    {#each graph.nodes as n (n.id)}
                        {@const label = labelPosition(n)}
                        <g
                                role="button"
                                tabindex="0"
                                aria-label={`${n.name}: ${currency.format(n.inflow)} in, ${currency.format(n.outflow)} out`}
                                class="cursor-pointer outline-none"
                                onmouseenter={() => (focus = n.id)}
                                onmouseleave={() => (focus = null)}
                                onfocus={() => (focus = n.id)}
                                onblur={() => (focus = null)}
                                onclick={() => goto(`/accounts/${n.id}`)}
                                onkeydown={(e) => e.key === "Enter" && goto(`/accounts/${n.id}`)}
                                opacity={!focus || focus === n.id || graph.edges.some((e) => (e.from.id === focus && e.to.id === n.id) || (e.to.id === focus && e.from.id === n.id)) ? 1 : 0.3}
                        >
                            <title>{n.name}: {currency.format(n.inflow)} in, {currency.format(n.outflow)} out</title>
                            <circle cx={n.x} cy={n.y} r={n.r} fill="#0f172a" stroke={color(n.type)} stroke-width="3" />
                            <circle cx={n.x} cy={n.y} r={n.r - 3} fill={color(n.type)} opacity="0.18" />
                            <text x={label.x} y={label.y} text-anchor={label.anchor} dominant-baseline="middle" class="fill-white text-[13px] font-semibold">
                                {n.name}
                            </text>
                            <text x={label.x} y={label.y + 16} text-anchor={label.anchor} dominant-baseline="middle" class="fill-slate-400 text-[12px] tabular-nums">
                                {currency.format(through(n))}
                            </text>
                        </g>
                    {/each}
                </svg>
            </div>
        </section>

        <!-- The same numbers as text, biggest first -->
        <section class="rounded-2xl border border-slate-800 bg-slate-900" aria-labelledby="flows-heading">
            <h2 id="flows-heading" class="px-6 pt-6 text-lg font-semibold text-white">Biggest movements</h2>
            <ul class="mt-3 divide-y divide-slate-800 border-t border-slate-800">
                {#each [...graph.edges].sort((a, b) => b.amount - a.amount).slice(0, 12) as edge (edge.id)}
                    <li class="flex items-center justify-between gap-4 px-6 py-3 text-sm">
                        <span class="min-w-0 truncate text-slate-300">
                            <span style:color={color(edge.from.type)}>{edge.from.name}</span>
                            <span class="text-slate-600">→</span>
                            <span style:color={color(edge.to.type)}>{edge.to.name}</span>
                        </span>
                        <span class="shrink-0 text-right tabular-nums">
                            <span class="font-semibold text-white">{currency.format(edge.amount)}</span>
                            <span class="block text-xs text-slate-500">{edge.transactions} {edge.transactions === 1 ? "transaction" : "transactions"}</span>
                        </span>
                    </li>
                {/each}
            </ul>
        </section>
    {:else}
        <p class="rounded-2xl border border-slate-800 bg-slate-900 p-12 text-center text-sm text-slate-400">
            No money moved {data.month ? `in ${formatMonth(`${data.month}-01`)}` : "yet"}. Import a statement to see where it goes.
        </p>
    {/if}
</div>
