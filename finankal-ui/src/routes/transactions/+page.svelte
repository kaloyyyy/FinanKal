<script lang="ts">
    import { formatDate } from "$lib/utils/date";
    import type { Transaction } from "$lib/server/api/transactions";
    import type { Account } from "$lib/server/api/accounts";
    import { currency } from "$lib/utils/currency.ts";
    import { enhance } from "$app/forms";
    import { ArrowDownLeft, ArrowUpRight, ArrowLeftRight, Receipt, Plus, MoveRight, Trash2, CircleCheck, CircleAlert, LoaderCircle } from "lucide-svelte";

    let { data, form } = $props();

    let clearDialog: HTMLDialogElement;
    let clearing = $state(false);

    const accountsById = $derived(new Map((data.accounts as Account[]).map((a) => [a.id, a])));
    const name = (id: string) => accountsById.get(id)?.name ?? "Unknown account";
    const isAsset = (id: string) => accountsById.get(id)?.type === "ASSET";

    type Direction = "IN" | "OUT" | "TRANSFER";

    // Money leaves the credited accounts and lands in the debited ones.
    function describe(t: Transaction) {
        const entries = t.entries ?? [];
        const debits = entries.filter((e) => e.type === "DEBIT");
        const credits = entries.filter((e) => e.type === "CREDIT");
        const amount = debits.reduce((sum, e) => sum + Number(e.amount), 0);

        const intoAsset = debits.some((e) => isAsset(e.accountId));
        const outOfAsset = credits.some((e) => isAsset(e.accountId));
        const direction: Direction = intoAsset && !outOfAsset ? "IN" : outOfAsset && !intoAsset ? "OUT" : "TRANSFER";

        return {
            amount,
            direction,
            from: [...new Set(credits.map((e) => name(e.accountId)))],
            to: [...new Set(debits.map((e) => name(e.accountId)))]
        };
    }

    const dayKey = (d: string) => new Date(d).toDateString();
    const today = new Date().toDateString();
    const yesterday = new Date(Date.now() - 864e5).toDateString();

    function dayLabel(key: string) {
        if (key === today) return "Today";
        if (key === yesterday) return "Yesterday";
        return formatDate(new Date(key));
    }

    const days = $derived.by(() => {
        const map = new Map<string, Transaction[]>();
        for (const t of (data.transactions ?? []) as Transaction[]) {
            const key = dayKey(t.createdAt);
            map.set(key, [...(map.get(key) ?? []), t]);
        }
        return [...map];
    });

    const style: Record<Direction, { icon: typeof Receipt; tone: string; sign: string }> = {
        IN: { icon: ArrowDownLeft, tone: "bg-emerald-500/10 text-emerald-400", sign: "+" },
        OUT: { icon: ArrowUpRight, tone: "bg-rose-500/10 text-rose-400", sign: "−" },
        TRANSFER: { icon: ArrowLeftRight, tone: "bg-slate-800 text-slate-300", sign: "" }
    };
    const amountTone: Record<Direction, string> = { IN: "text-emerald-400", OUT: "text-white", TRANSFER: "text-slate-300" };

    const time = (d: string) => new Date(d).toLocaleTimeString("en-PH", { hour: "numeric", minute: "2-digit" });
</script>

<svelte:head>
    <title>Transactions | FinanKal</title>
</svelte:head>

<div class="space-y-8">
    <header class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight text-white">Transactions</h1>
            <p class="mt-1 text-slate-400">
                {#if data.total}
                    {data.total} recorded. Each one shows where the money came from and where it went.
                {:else}
                    Where your money came from and where it went.
                {/if}
            </p>
        </div>
        <div class="flex flex-col-reverse gap-3 sm:flex-row">
            {#if data.total}
                <button type="button"
                        onclick={() => clearDialog.showModal()}
                        class="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-700 bg-slate-900 px-5 py-3 font-semibold text-slate-300 transition-colors hover:border-red-500/40 hover:text-red-300 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500">
                    <Trash2 class="h-4 w-4" />
                    Clear all
                </button>
            {/if}
            <a href="/transactions/new"
               class="inline-flex items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-3 font-semibold text-white transition-colors hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500">
                <Plus class="h-4 w-4" />
                New transaction
            </a>
        </div>
    </header>

    {#if form?.cleared !== undefined}
        <div role="status" class="flex items-start gap-3 rounded-xl border border-emerald-500/20 bg-emerald-500/10 p-4 text-sm text-emerald-300">
            <CircleCheck class="mt-0.5 h-4 w-4 shrink-0 text-emerald-400" />
            Cleared {form.cleared} {form.cleared === 1 ? "transaction" : "transactions"}. Your accounts are still there, with balances back to zero.
        </div>
    {:else if form?.clearError}
        <div role="alert" class="flex items-start gap-3 rounded-xl border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-300">
            <CircleAlert class="mt-0.5 h-4 w-4 shrink-0 text-red-400" />
            {form.clearError}
        </div>
    {/if}

    <dialog bind:this={clearDialog}
            aria-labelledby="clear-title"
            aria-describedby="clear-description"
            class="m-auto w-[calc(100%-2rem)] max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-0 text-white shadow-2xl backdrop:bg-black/70">
        <form method="POST"
              action="?/clear"
              class="p-6"
              use:enhance={() => {
                  clearing = true;
                  return async ({ update }) => {
                      await update();
                      clearing = false;
                      clearDialog.close();
                  };
              }}>
            <h2 id="clear-title" class="text-lg font-semibold">
                Clear all {data.total} {data.total === 1 ? "transaction" : "transactions"}?
            </h2>
            <div id="clear-description" class="mt-2 space-y-2 text-sm leading-6 text-slate-400">
                <p>
                    Every transaction is deleted, along with the credit card statements built from them.
                    This can't be undone.
                </p>
                <p>Your accounts and credit cards are kept, with their balances back to zero.</p>
            </div>

            <div class="mt-6 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
                <button type="button"
                        onclick={() => clearDialog.close()}
                        class="rounded-xl border border-slate-700 bg-slate-800 px-4 py-2.5 text-sm font-semibold text-slate-200 transition-colors hover:border-slate-600 hover:text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500">
                    Keep transactions
                </button>
                <button type="submit"
                        disabled={clearing}
                        class="inline-flex items-center justify-center gap-2 rounded-xl bg-red-600 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-red-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-500 disabled:opacity-60">
                    {#if clearing}
                        <LoaderCircle class="h-4 w-4 animate-spin motion-reduce:animate-none" />
                        Clearing…
                    {:else}
                        Clear transactions
                    {/if}
                </button>
            </div>
        </form>
    </dialog>

    {#if days.length === 0}
        <div class="rounded-2xl border border-dashed border-slate-700 px-6 py-16 text-center">
            <Receipt class="mx-auto h-8 w-8 text-slate-500" />
            <h2 class="mt-4 text-lg font-semibold text-white">No transactions yet</h2>
            <p class="mx-auto mt-2 max-w-sm text-sm text-slate-400">
                Record a payment, a purchase or a transfer between accounts, or import a bank statement.
            </p>
            <a href="/transactions/new"
               class="mt-6 inline-flex items-center gap-2 rounded-xl bg-blue-600 px-5 py-3 font-medium text-white transition-colors hover:bg-blue-500">
                <Plus class="h-4 w-4" />
                New transaction
            </a>
        </div>
    {:else}
        {#each days as [key, transactions], i (key)}
            <section aria-labelledby={`day-${i}`}>
                <h2 id={`day-${i}`}
                    class="sticky top-0 z-10 -mx-3 border-b border-slate-800 bg-slate-950/90 px-3 py-2 text-sm font-semibold text-slate-300 backdrop-blur">
                    {dayLabel(key)}
                </h2>

                <ul class="divide-y divide-slate-800/70">
                    {#each transactions as t (t.id)}
                        {@const d = describe(t)}
                        {@const s = style[d.direction]}
                        <li>
                            <a href={`/transactions/${t.id}`}
                               class="-mx-3 grid grid-cols-[2.5rem_1fr_auto] items-center gap-x-4 rounded-xl px-3 py-4 transition-colors hover:bg-slate-900 focus-visible:bg-slate-900 focus-visible:outline-2 focus-visible:outline-blue-500">
                                <span class={`flex h-10 w-10 items-center justify-center rounded-full ${s.tone}`}>
                                    <s.icon class="h-4 w-4" />
                                </span>

                                <span class="min-w-0">
                                    <span class="block truncate font-medium text-white">
                                        {t.description || "Untitled transaction"}
                                    </span>
                                    {#if d.from.length || d.to.length}
                                        <span class="mt-1 flex min-w-0 items-center gap-1.5 text-sm text-slate-400">
                                            <span class="truncate">{d.from.join(", ") || "—"}</span>
                                            <MoveRight class="h-3.5 w-3.5 shrink-0 text-slate-600" aria-label="to" />
                                            <span class="truncate text-slate-300">{d.to.join(", ") || "—"}</span>
                                        </span>
                                    {:else}
                                        <span class="mt-1 block text-sm text-slate-500">No entries</span>
                                    {/if}
                                </span>

                                <span class="text-right">
                                    <span class={`block font-semibold tabular-nums ${amountTone[d.direction]}`}>
                                        {s.sign}{currency.format(d.amount)}
                                    </span>
                                    <span class="mt-1 block text-xs text-slate-500">{time(t.createdAt)}</span>
                                </span>
                            </a>
                        </li>
                    {/each}
                </ul>
            </section>
        {/each}

        {#if data.total > (data.transactions?.length ?? 0)}
            <p class="text-center text-sm text-slate-500">
                Showing the latest {data.transactions.length} of {data.total}.
            </p>
        {/if}
    {/if}
</div>
