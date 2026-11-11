<script lang="ts">
    import { formatDate, formatMonth } from "$lib/utils/date";
    import type {LedgerEntryDto} from "$lib/types/ledger";
    import {currency} from "$lib/utils/currency.ts";


    interface Props {
        entries: LedgerEntryDto[];
        billingDay?: number;
        statementShift?: string;
    }

    let {entries, billingDay, statementShift}: Props = $props();
    const byStatement = $derived(Number(billingDay) > 0);

    interface Month {
        label: string; // "NOV 2026"
        key: string;
        start: number;
        end: number;
        spent: number;
        paid: number;
        entries: LedgerEntryDto[];
    }

    function statementDateFor(value: string | undefined): string {
        const date = new Date(value ?? "");
        if (Number.isNaN(date.getTime())) return "";
        const parts = manilaDateParts(date);
        const month = Number(parts.month) - 1;
        const entryDate = `${parts.year}-${parts.month}-${parts.day}`;
        const close = statementClose(Number(parts.year), month);
        return entryDate <= close ? close : statementClose(Number(parts.year), month + 1);
    }

    function manilaDateParts(date: Date) {
        return Object.fromEntries(new Intl.DateTimeFormat("en-CA", {
            timeZone: "Asia/Manila", year: "numeric", month: "2-digit", day: "2-digit"
        }).formatToParts(date).map((part) => [part.type, part.value]));
    }

    function calendarMonthKey(value: string | undefined): string {
        const date = new Date(value ?? "");
        if (Number.isNaN(date.getTime())) return "";
        const parts = manilaDateParts(date);
        return `${parts.year}-${parts.month}`;
    }

    function statementClose(year: number, month: number): string {
        const last = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
        const close = new Date(Date.UTC(year, month, Math.min(Number(billingDay), last)));
        if (statementShift === "NEXT_BUSINESS_DAY") {
            while (close.getUTCDay() === 0 || close.getUTCDay() === 6) close.setUTCDate(close.getUTCDate() + 1);
        } else if (statementShift === "NEAREST_BUSINESS_DAY") {
            if (close.getUTCDay() === 6) close.setUTCDate(close.getUTCDate() - 1);
            else if (close.getUTCDay() === 0) close.setUTCDate(close.getUTCDate() + 1);
        }
        return close.toISOString().slice(0, 10);
    }

    // Entries grouped newest first, with the balance before and after each group: by statement cycle for
    // a card, like the bank lists them (a payment sits in the cycle it was made in), else by month.
    // The engine's balance is debits − credits for every account type, so it's the same here.
    const months = $derived.by(() => {
        const oldestFirst = [...entries].sort(
            (a, b) => Date.parse(a.createdAt ?? "") - Date.parse(b.createdAt ?? "") || (a.transactionId ?? "").localeCompare(b.transactionId ?? "")
        );
        const result: Month[] = [];
        let balance = 0;
        for (const entry of oldestFirst) {
            const key = byStatement ? statementDateFor(entry.createdAt) : calendarMonthKey(entry.createdAt);
            let month = result.at(-1);
            if (month?.key !== key) {
                const label = byStatement ? `Statement ending ${formatDate(key)}` : formatMonth(`${key}-01`);
                month = { key, label, start: balance, end: balance, spent: 0, paid: 0, entries: [] };
                result.push(month);
            }
            const amount = Number(entry.amount);
            const change = (entry.type === "DEBIT" ? 1 : -1) * amount;
            balance += change;
            month.end = balance;
            // Statement imports date the account's opening balance on its first day: it's where the group starts.
            // Name must match openingBalanceName in engine-go/internal/statement/service.go.
            if (entry.description === "Opening balance") month.start += change;
            else if (entry.type === "CREDIT") month.spent = Math.round((month.spent + amount) * 100) / 100;
            else month.paid = Math.round((month.paid + amount) * 100) / 100;
            month.entries.unshift(entry); // newest first within the group
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
                            <span class="flex flex-wrap gap-x-4 gap-y-1 text-xs font-normal text-slate-400 tabular-nums">
                                <span>Start <span class="font-semibold text-slate-200">{currency.format(month.start)}</span></span>
                                {#if byStatement}
                                    <span>Spent <span class="font-semibold text-rose-300">{currency.format(month.spent)}</span></span>
                                    <span>Paid <span class="font-semibold text-emerald-300">{currency.format(month.paid)}</span></span>
                                {/if}
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
                        {#if entry.transactionId}
                            <!-- Opens the transaction to view, edit or delete it; "back" returns to this account. -->
                            <a href={`/transactions/${entry.transactionId}?from=${encodeURIComponent(`/accounts/${entry.accountId}`)}`}
                               class="font-medium text-white underline-offset-4 transition hover:text-blue-400 hover:underline focus-visible:outline-2 focus-visible:outline-blue-500">
                                {entry.description || "No description"}
                            </a>
                        {:else}
                            <p class="font-medium text-white">
                                {entry.description || "No description"}
                            </p>
                        {/if}
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
