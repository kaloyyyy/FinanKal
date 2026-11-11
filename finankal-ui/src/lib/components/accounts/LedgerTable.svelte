<script lang="ts">
    import { formatDate, formatMonth } from "$lib/utils/date";
    import type {LedgerEntryDto} from "$lib/types/ledger";
    import {currency} from "$lib/utils/currency.ts";

    const roundMoney = (amount: number) => Math.round(amount * 100) / 100;
    const oldestFirst = (a: LedgerEntryDto, b: LedgerEntryDto) =>
        Date.parse(a.createdAt ?? "") - Date.parse(b.createdAt ?? "") || (a.transactionId ?? "").localeCompare(b.transactionId ?? "");
    const newestFirst = (a: LedgerEntryDto, b: LedgerEntryDto) => oldestFirst(b, a);

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

    function statementGroup(entry: LedgerEntryDto): string {
        const nextClose = statementDateFor(entry.createdAt);
        // Card repayments are DEBIT entries: debits increase the card's signed balance toward zero.
        if (entry.type !== "DEBIT" || entry.description === "Opening balance" || !nextClose) return nextClose;
        const parts = manilaDateParts(new Date(entry.createdAt ?? ""));
        const entryDate = `${parts.year}-${parts.month}-${parts.day}`;
        const year = Number(parts.year), month = Number(parts.month) - 1;
        const thisClose = statementClose(year, month);
        if (entryDate >= thisClose) return thisClose;
        return statementClose(year, month - 1);
    }

    // Card payments settle closed statements oldest-first; any excess rolls forward to the next statement.
    const months = $derived.by(() => {
        const chronologicalEntries = [...entries].sort(oldestFirst);
        type Group = Month & { openingBalance: number; payments: LedgerEntryDto[] };
        const grouped = new Map<string, Group>();
        const groupFor = (key: string): Group => {
            let group = grouped.get(key);
            if (!group) {
                group = { key, label: byStatement ? `Statement ending ${formatDate(key)}` : formatMonth(`${key}-01`), start: 0, end: 0, spent: 0, paid: 0, entries: [], openingBalance: 0, payments: [] };
                grouped.set(key, group);
            }
            return group;
        };
        for (const entry of chronologicalEntries) {
            const key = byStatement
                ? entry.type === "DEBIT" && entry.description !== "Opening balance"
                    ? statementGroup(entry)
                    : statementDateFor(entry.createdAt)
                : calendarMonthKey(entry.createdAt);
            const group = groupFor(key);
            const amount = Number(entry.amount);
            if (byStatement && entry.type === "DEBIT" && entry.description !== "Opening balance") {
                group.payments.push(entry);
            } else {
                if (!byStatement) group.end += (entry.type === "DEBIT" ? 1 : -1) * amount;
                if (entry.description === "Opening balance") group.openingBalance += (entry.type === "DEBIT" ? 1 : -1) * amount;
                else if (byStatement && entry.type === "CREDIT") group.spent = roundMoney(group.spent + amount);
                group.entries.push(entry);
            }
        }

        const result = [...grouped.values()].sort((a, b) => a.key.localeCompare(b.key));
        if (!byStatement) {
            let monthBalance = 0;
            for (const group of result) {
                const change = group.end;
                group.start = monthBalance + group.openingBalance;
                monthBalance += change;
                group.end = monthBalance;
                group.entries.sort(newestFirst);
            }
            return result.reverse();
        }
        let balance = 0;
        let carriedPayments: { entry: LedgerEntryDto; amount: number }[] = [];
        for (const group of result) {
            group.start = balance + group.openingBalance;
            const dueBeforePayments = group.start - group.spent;
            const paymentQueue = [
                ...carriedPayments,
                ...group.payments.sort(oldestFirst).map((entry) => ({ entry, amount: Number(entry.amount) }))
            ];
            carriedPayments = [];
            let remainingDue = Math.max(0, -dueBeforePayments);
            for (const payment of paymentQueue) {
                const applied = roundMoney(Math.min(payment.amount, remainingDue));
                if (applied > 0) {
                    const description = payment.amount < Number(payment.entry.amount) ? `${payment.entry.description ?? "Payment"} · statement allocation` : payment.entry.description;
                    group.entries.push({ ...payment.entry, amount: applied, description });
                    group.paid = roundMoney(group.paid + applied);
                    remainingDue = roundMoney(remainingDue - applied);
                }
                if (payment.amount > applied) carriedPayments.push({ entry: payment.entry, amount: roundMoney(payment.amount - applied) });
            }
            group.end = dueBeforePayments + group.paid;
            balance = group.end;
            group.entries.sort(newestFirst);
        }

        if (byStatement && carriedPayments.length && result.length) {
            const last = result[result.length - 1];
            const nextDate = new Date(`${last.key}T00:00:00Z`);
            nextDate.setUTCDate(nextDate.getUTCDate() + 1);
            const nextKey = statementDateFor(nextDate.toISOString().slice(0, 10));
            const future = groupFor(nextKey);
            future.start = balance;
            future.paid = carriedPayments.reduce((sum, payment) => sum + payment.amount, 0);
            future.end = balance + future.paid;
            future.entries = carriedPayments.map(({ entry, amount }) => ({
                ...entry, amount,
                description: amount < Number(entry.amount) ? `${entry.description ?? "Payment"} · statement allocation` : entry.description
            }));
            future.entries.sort(newestFirst);
            result.push(future);
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
                                <span>{byStatement ? "After payments" : "End"} <span class="font-semibold text-slate-200">{currency.format(month.end)}</span></span>
                                {#if byStatement}
                                    <span class={Math.abs(month.end) < 0.005 ? "font-semibold text-emerald-300" : month.end < 0 ? "font-semibold text-amber-300" : "font-semibold text-blue-300"}>
                                        {Math.abs(month.end) < 0.005 ? "Paid" : month.end < 0 ? `Still owed ${currency.format(-month.end)}` : `Credit balance ${currency.format(month.end)}`}
                                    </span>
                                {/if}
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
