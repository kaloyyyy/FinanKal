<script lang="ts">
    import { enhance } from "$app/forms";
    import { Plus, X } from "lucide-svelte";
    import { currency } from "$lib/utils/currency";
    import { formatDate } from "$lib/utils/date";

    interface Props {
        accountName: string;
        payFrom: { id: string; name: string }[];
        error?: string;
    }

    let { accountName, payFrom, error }: Props = $props();

    let dialog: HTMLDialogElement;
    let saving = $state(false);

    let sourceAccountId = $state("");
    // Number inputs bind numbers (null when empty).
    let total = $state<number | null>(null);
    let count = $state<number | null>(12);
    let firstDue = $state("");
    let customName = $state<string | null>(null);
    let rows = $state<{ dueDate: string; amount: number | null }[]>([]);

    export function open() {
        dialog.showModal();
    }

    const toCents = (amount: number | null) => Math.round(Number(amount) * 100);

    /** yyyy-MM-dd, `months` after `first`, on the same day or the month's last day if shorter. */
    function addMonths(first: string, months: number): string {
        const [y, m, d] = first.split("-").map(Number);
        const lastDay = new Date(Date.UTC(y, m + months, 0)).getUTCDate();
        return new Date(Date.UTC(y, m - 1 + months, Math.min(d, lastDay))).toISOString().slice(0, 10);
    }

    /** Spreads the total over monthly installments; leftover centavos go to the earliest ones (943.37, 943.37, 943.36). */
    function fill() {
        const n = Math.floor(Number(count));
        const cents = toCents(total);
        if (!firstDue || !(n >= 1 && n <= 360) || !(cents >= n)) return;
        const base = Math.floor(cents / n);
        const extra = cents % n;
        rows = Array.from({ length: n }, (_, i) => ({
            dueDate: addMonths(firstDue, i),
            amount: (base + (i < extra ? 1 : 0)) / 100
        }));
    }

    function addRow() {
        const last = rows.at(-1);
        rows.push({ dueDate: last?.dueDate ? addMonths(last.dueDate, 1) : firstDue, amount: last?.amount ?? null });
    }

    const sum = $derived(rows.reduce((cents, r) => cents + (toCents(r.amount) || 0), 0) / 100);
    const off = $derived(total && rows.length ? (toCents(total) - Math.round(sum * 100)) / 100 : 0);

    const suggestedName = $derived(
        `${accountName} ${currency.format(Number(total) || sum).replace(".00", "")} (${rows.length} ${rows.length === 1 ? "installment" : "installments"}${rows[0]?.dueDate ? `, from ${formatDate(rows[0].dueDate)}` : ""})`
    );
    const name = $derived(customName ?? suggestedName);

    // The first thing stopping the schedule from being saved, in the order the form reads.
    const problem = $derived.by(() => {
        if (!sourceAccountId) return "Choose the account you pay from.";
        if (!rows.length) return "Fill in the total, count and first due date, or add installments one by one.";
        for (const [i, r] of rows.entries()) {
            if (!r.dueDate) return `Installment ${i + 1} needs a due date.`;
            if (!(toCents(r.amount) > 0)) return `Installment ${i + 1} needs an amount above zero.`;
            if (i > 0 && r.dueDate <= rows[i - 1].dueDate) return `Installment ${i + 1} must be due after installment ${i}.`;
        }
        if (!name.trim()) return "Give the schedule a name.";
        return null;
    });

    const payload = $derived(JSON.stringify({ name: name.trim(), sourceAccountId, installments: rows }));

    function reset() {
        total = null;
        count = 12;
        firstDue = "";
        customName = null;
        rows = [];
    }

    const field =
        "w-full rounded-xl border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-white tabular-nums [color-scheme:dark] placeholder:text-slate-600 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-500/30";
</script>

<dialog
        bind:this={dialog}
        aria-labelledby="installments-title"
        class="m-auto max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-2xl overflow-hidden rounded-2xl border border-slate-800 bg-slate-900 p-0 text-slate-100 shadow-2xl backdrop:bg-slate-950/75 backdrop:backdrop-blur-sm open:flex open:flex-col"
>
    <form
            method="POST"
            action="?/addInstallments"
            class="flex min-h-0 flex-col"
            use:enhance={() => {
                saving = true;
                return async ({ result, update }) => {
                    saving = false;
                    if (result.type === "success") {
                        dialog.close();
                        reset();
                    }
                    await update({ reset: false });
                };
            }}
    >
        <header class="flex items-start justify-between gap-4 border-b border-slate-800 px-6 py-5">
            <div>
                <h2 id="installments-title" class="text-lg font-semibold text-white">Add installments</h2>
                <p class="mt-1 text-sm text-slate-400">
                    Each one shows under {accountName}'s repayment until it's paid.
                </p>
            </div>
            <button
                    type="button"
                    onclick={() => dialog.close()}
                    class="rounded-lg p-1.5 text-slate-400 transition hover:bg-slate-800 hover:text-white focus-visible:outline-2 focus-visible:outline-blue-500"
                    aria-label="Close"
            >
                <X class="h-5 w-5" />
            </button>
        </header>

        <div class="min-h-0 flex-1 space-y-6 overflow-y-auto px-6 py-5">
            <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-slate-300">Pay from</span>
                <select bind:value={sourceAccountId} class={field} required>
                    <option value="" disabled>Choose a bank or wallet</option>
                    {#each payFrom as a (a.id)}
                        <option value={a.id}>{a.name}</option>
                    {/each}
                </select>
            </label>

            <fieldset>
                <legend class="mb-1.5 text-sm font-medium text-slate-300">Spread a total evenly</legend>
                <div class="grid gap-3 sm:grid-cols-[1.4fr_1fr_1.3fr]">
                    <label class="block">
                        <span class="mb-1 block text-xs text-slate-400">Total to repay (₱)</span>
                        <input type="number" inputmode="decimal" min="0" step="0.01" placeholder="33,000.00"
                               bind:value={total} oninput={fill} class={field} />
                    </label>
                    <label class="block">
                        <span class="mb-1 block text-xs text-slate-400">Installments</span>
                        <input type="number" inputmode="numeric" min="1" max="360" step="1"
                               bind:value={count} oninput={fill} class={field} />
                    </label>
                    <label class="block">
                        <span class="mb-1 block text-xs text-slate-400">First due</span>
                        <input type="date" bind:value={firstDue} oninput={fill} class={field} />
                    </label>
                </div>
                <p class="mt-2 text-xs text-slate-500">Monthly on the same day. Edit any row below for odd amounts or dates.</p>
            </fieldset>

            <section aria-label="Schedule">
                {#if rows.length}
                    <div class="grid grid-cols-[2rem_1fr_1fr_2rem] items-center gap-x-2 border-b border-slate-800 pb-2 text-xs text-slate-500">
                        <span>#</span><span>Due</span><span>Amount (₱)</span><span></span>
                    </div>
                    <ol class="divide-y divide-slate-800/70">
                        {#each rows as row, i}
                            <li class="grid grid-cols-[2rem_1fr_1fr_2rem] items-center gap-x-2 py-1.5">
                                <span class="text-sm text-slate-500 tabular-nums">{i + 1}</span>
                                <input type="date" bind:value={row.dueDate} class={field}
                                       aria-label={`Installment ${i + 1} due date`} />
                                <input type="number" inputmode="decimal" min="0" step="0.01" bind:value={row.amount}
                                       class={`${field} text-right`} aria-label={`Installment ${i + 1} amount`} />
                                <button type="button" onclick={() => rows.splice(i, 1)}
                                        class="justify-self-center rounded-lg p-1 text-slate-500 transition hover:bg-slate-800 hover:text-red-400 focus-visible:outline-2 focus-visible:outline-blue-500"
                                        aria-label={`Remove installment ${i + 1}`}>
                                    <X class="h-4 w-4" />
                                </button>
                            </li>
                        {/each}
                    </ol>
                {/if}
                <button type="button" onclick={addRow}
                        class="mt-2 inline-flex items-center gap-1.5 rounded-xl px-2 py-1.5 text-sm font-medium text-blue-400 transition hover:bg-slate-800 hover:text-blue-300 focus-visible:outline-2 focus-visible:outline-blue-500">
                    <Plus class="h-4 w-4" /> Add installment
                </button>
            </section>

            <label class="block">
                <span class="mb-1.5 block text-sm font-medium text-slate-300">Name</span>
                <input type="text" value={name} oninput={(e) => (customName = e.currentTarget.value)}
                       maxlength="255" class={field} />
                <span class="mt-1 block text-xs text-slate-500">How the installments are labelled in deadlines.</span>
            </label>
        </div>

        <footer class="space-y-3 border-t border-slate-800 bg-slate-950/40 px-6 py-4">
            {#if error}
                <p role="alert" class="rounded-xl bg-red-500/10 px-3 py-2 text-sm text-red-400">
                    {error.charAt(0).toUpperCase() + error.slice(1)}
                </p>
            {/if}
            <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="text-sm" aria-live="polite">
                    <p class="font-semibold text-white tabular-nums">{currency.format(sum)}
                        <span class="font-normal text-slate-400">in {rows.length} {rows.length === 1 ? "installment" : "installments"}</span>
                    </p>
                    {#if off}
                        <p class="text-amber-300 tabular-nums">{currency.format(Math.abs(off))} {off > 0 ? "short of" : "over"} the total</p>
                    {/if}
                    {#if problem}
                        <p class="text-slate-500">{problem}</p>
                    {/if}
                </div>
                <div class="flex gap-2">
                    <button type="button" onclick={() => dialog.close()}
                            class="rounded-xl border border-slate-700 bg-slate-800 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-blue-500">
                        Cancel
                    </button>
                    <button type="submit" disabled={!!problem || saving}
                            class="rounded-xl bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 disabled:cursor-not-allowed disabled:opacity-40">
                        {saving ? "Adding…" : `Add ${rows.length || ""} ${rows.length === 1 ? "installment" : "installments"}`}
                    </button>
                </div>
            </div>
        </footer>

        <input type="hidden" name="schedule" value={payload} />
    </form>
</dialog>
