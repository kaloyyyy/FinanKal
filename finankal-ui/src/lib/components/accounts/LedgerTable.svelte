<script lang="ts">
    import type {LedgerEntryDto} from "$lib/types/ledger";
    import {currency} from "$lib/utils/currency.ts";


    interface Props {
        entries: LedgerEntryDto[];
    }

    let {entries}: Props = $props();
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
                    Account
                </th>

                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Type
                </th>

                <th class="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Amount
                </th>
            </tr>
            </thead>

            <tbody class="divide-y divide-slate-800">

            {#each entries as entry}

                <tr class="transition hover:bg-slate-800/50">

                    <td class="whitespace-nowrap px-5 py-4 text-sm text-slate-400">
                        {new Date(entry.createdAt).toLocaleDateString()}
                    </td>

                    <td class="px-5 py-4">
                        <p class="font-medium text-white">
                            {entry.description || "No description"}
                        </p>
                    </td>

                    <td class="px-5 py-4">
                            <span class="font-mono text-xs text-slate-400">
                                {entry.accountId}
                            </span>
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

            {:else}

                <tr>
                    <td
                            colspan="5"
                            class="px-5 py-12 text-center text-sm text-slate-500"
                    >
                        No ledger entries found.
                    </td>
                </tr>

            {/each}

            </tbody>

        </table>
    </div>

</div>