<script lang="ts">
    import { currency } from "$lib/utils/currency";
    import { coverage, type PositionTile, type Slice } from "$lib/utils/payables";
    import { bankBackground } from "$lib/utils/bankColors";

    let {
        tile,
        slices = {}
    }: {
        tile: PositionTile;
        /** Per term index, the accounts that make it up; a term without slices is one line. */
        slices?: Record<number, Slice[]>;
    } = $props();

    // The sum as a tally, one line per account where known, like adding it up on paper.
    const lines = $derived(
        tile.terms.flatMap((term, i) => {
            const minus = term.sign === "−";
            const parts = slices[i]?.length
                ? slices[i]
                      .filter((s) => s.amount !== 0)
                      .map((s) => ({
                          id: s.id,
                          label: s.label ?? s.name,
                          amount: s.amount,
                          background: bankBackground(s.name, minus ? "owed" : s.predicted ? "predicted" : "held")
                      }))
                : [{ id: `term-${i}`, label: term.label, amount: term.amount, background: null }];
            return parts.map((p) => ({ ...p, sign: term.sign }));
        })
    );

    const meter = $derived(coverage(tile));
</script>

<article class="flex h-full flex-col rounded-2xl border border-slate-800 bg-slate-900/60 p-5 sm:p-6">
    <h3 class="text-base font-semibold text-white">{tile.question}</h3>
    <p class="mt-0.5 text-sm text-slate-500">{tile.when}</p>

    <p class={`mt-4 text-4xl font-bold tracking-tight tabular-nums ${tile.value < 0 ? "text-rose-300" : "text-white"}`}>
        {currency.format(tile.value)}
    </p>

    <!-- Can it be covered? Fills by the share of what's going out that's covered. -->
    <div
            class="mt-4 h-1.5 overflow-hidden rounded-full bg-slate-800"
            role="meter"
            aria-valuemin={0}
            aria-valuemax={100}
            aria-valuenow={Math.round(meter.fill * 100)}
            aria-label={meter.text}
    >
        <div
                class={`h-full rounded-full ${meter.covered ? "bg-emerald-400" : "bg-rose-400"}`}
                style:width={`${Math.max(meter.fill * 100, meter.fill > 0 ? 1.5 : 0)}%`}
        ></div>
    </div>
    <p class={`mt-2 text-sm ${meter.covered ? "text-emerald-300" : "text-rose-300"}`}>{meter.text}</p>

    <!-- The working: each account, added or taken away, down to the total -->
    <table class="mt-6 w-full text-sm">
        <caption class="sr-only">How {currency.format(tile.value)} adds up</caption>
        <tbody>
            {#each lines as line (line.id)}
                <tr>
                    <td class="w-4 py-1 align-top text-slate-500" aria-hidden="true">{line.sign}</td>
                    <th scope="row" class="py-1 pr-3 text-left font-normal text-slate-300">
                        <span class="flex items-center gap-2">
                            <span
                                    aria-hidden="true"
                                    class={`h-3 w-3 shrink-0 rounded-sm ${line.background ? "" : line.sign === "−" ? "bg-rose-500" : "bg-emerald-500"}`}
                                    style:background={line.background}
                            ></span>
                            <span class="min-w-0">
                                <span class="sr-only">{line.sign === "−" ? "minus " : line.sign === "+" ? "plus " : ""}</span>{line.label}
                            </span>
                        </span>
                    </th>
                    <td class={`py-1 text-right tabular-nums ${line.sign === "−" ? "text-rose-300" : line.sign === "+" ? "text-emerald-300" : "text-slate-200"}`}>
                        {currency.format(line.amount)}
                    </td>
                </tr>
            {/each}
        </tbody>
        <tfoot>
            <tr class="border-t border-slate-700">
                <td class="pt-2" aria-hidden="true"></td>
                <th scope="row" class="pt-2 text-left font-medium text-slate-300">Total</th>
                <td class={`pt-2 text-right font-semibold tabular-nums ${tile.value < 0 ? "text-rose-300" : "text-white"}`}>
                    {currency.format(tile.value)}
                </td>
            </tr>
        </tfoot>
    </table>
</article>
