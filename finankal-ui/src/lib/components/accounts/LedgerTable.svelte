<script lang="ts">
    import { formatDate, formatMonth } from "$lib/utils/date";
    import type {LedgerEntryDto} from "$lib/types/ledger";
    import {currency} from "$lib/utils/currency.ts";


    interface Props {
        entries: LedgerEntryDto[];
    }

    let {entries}: Props = $props();

    interface Month {
        label: string; // "NOV 2026"
        start: number;
        end: number;
        entries: LedgerEntryDto[];
    }

    // Entries grouped by month, newest first, with the balance before and after each month.
    // The engine's balance is debits − credits for every account type, so it's the same here.
    const months = $derived.by(() => {
        const oldestFirst = [...entries].sort(
            (a, b) => Date.parse(a.createdAt ?? "") - Date.parse(b.createdAt ?? "")
        );
        const result: Month[] = [];
        let balance = 0;
        for (const entry of oldestFirst) {
            const label = formatMonth(entry.createdAt);
            let month = result.at(-1);
            if (month?.label !== label) {
                month = { label, start: balance, end: balance, entries: [] };
                result.push(month);
            }
            const change = (entry.type === "DEBIT" ? 1 : -1) * Number(entry.amount);
            balance += change;
            month.end = balance;
            // Statement imports date the account's opening balance on its first day: it's where the month starts.
            // Name must match openingBalanceName in engine-go/internal/statement/service.go.
            if (entry.description === "Opening balance") {
                month.start += change;
            }
            month.entries.unshift(entry); // newest first within the month
        }
        return result.reverse();
    });
</script>

<div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">

    <div class="overflow-x-auto">
        <table class="w-full text-left">

            <thead class="border-b border-slate-800 bg-slate-950/50">
            <tr>
                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Date
                </th>

                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Description
                </th>

                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Type
                </th>

                <th class="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Amount
                </th>
            </tr>
            </thead>

            {#each months as month}
            <tbody class="divide-y divide-slate-800 border-b border-slate-800">

                <tr class="bg-slate-950/40">
                    <th colspan="4" scope="rowgroup" class="px-5 py-3 text-left">
                        <div class="flex flex-wrap items-baseline justify-between gap-x-6 gap-y-1">
                            <span class="text-sm font-semibold tracking-wide text-white">{month.label}</span>
                            <span class="flex gap-6 text-xs font-normal text-slate-400 tabular-nums">
                                <span>Start <span class="font-semibold text-slate-200">{currency.format(month.start)}</span></span>
                                <span>End <span class="font-semibold text-slate-200">{currency.format(month.end)}</span></span>
                            </span>
                        </div>
                    </th>
                </tr>

            {#each month.entries as entry}

                <tr class="transition hover:bg-slate-800/50">

                    <td class="whitespace-nowrap px-5 py-4 text-sm text-slate-400">
                        {formatDate(entry.createdAt)}
                    </td>

                    <td class="px-5 py-4">
                        <p class="font-medium text-white">
                            {entry.description || "No description"}
                        </p>
                    </td>

                    <td class="px-5 py-4">

                        {#if entry.type === "DEBIT"}

                                <span class="inline-flex rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-400">
                                    Debit
                                </span>

                        {:else}

                                <span class="inline-flex rounded-full bg-blue-500/10 px-3 py-1 text-xs font-medium text-blue-400">
                                    Credit
                                </span>

                        {/if}

                    </td>

                    <td
                            class={`whitespace-nowrap px-5 py-4 text-right font-semibold ${
                                entry.type === "DEBIT"
                                    ? "text-emerald-400"
                                    : "text-blue-400"
                            }`}
                    >
                        {currency.format(Number(entry.amount))}
                    </td>

                </tr>

            {/each}

            </tbody>
            {:else}
            <tbody>
                <tr>
                    <td
                            colspan="4"
                            class="px-5 py-12 text-center text-sm text-slate-500"
                    >
                        No ledger entries found.
                    </td>
                </tr>
            </tbody>
            {/each}

        </table>
    </div>

</div>