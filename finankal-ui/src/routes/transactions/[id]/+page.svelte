<script lang="ts">
    import { enhance } from "$app/forms";
    import { formatDate as formatDay } from "$lib/utils/date";
    let { data, form } = $props();

    const currency = new Intl.NumberFormat("en-PH", {
        style: "currency",
        currency: "PHP"
    });

    function formatDate(date: string | null | undefined) {
        if (!date) return "—";

        const time = new Date(date).toLocaleTimeString("en-PH", {
            hour: "numeric",
            minute: "2-digit",
            timeZone: "Asia/Manila"
        });
        return `${formatDay(date)}, ${time}`;
    }

    function typeClass(type: string) {
        return type === "DEBIT"
            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
            : "bg-blue-500/10 text-blue-400 border-blue-500/20";
    }

    let totalDebit = $derived(
        data.transaction.entries
            ?.filter((entry: any) => entry.type === "DEBIT")
            .reduce((sum: number, entry: any) => sum + Number(entry.amount), 0) ?? 0
    );

    let totalCredit = $derived(
        data.transaction.entries
            ?.filter((entry: any) => entry.type === "CREDIT")
            .reduce((sum: number, entry: any) => sum + Number(entry.amount), 0) ?? 0
    );

    let balanced = $derived(
        Math.abs(totalDebit - totalCredit) < 0.01
    );

    // ---- Editing: description, day, and each entry's account and amount (entries can't be added or removed).
    type Draft = { description: string; date: string; entries: { id: string; accountId: string; amount: number; type: "DEBIT" | "CREDIT" }[] };
    let editing = $state(false);
    let saving = $state(false);
    let draft = $state<Draft>({ description: "", date: "", entries: [] });

    function startEditing() {
        draft = {
            description: data.transaction.description ?? "",
            // The day as it reads in the Philippines (en-CA formats as yyyy-MM-dd).
            date: new Date(data.transaction.createdAt).toLocaleDateString("en-CA", { timeZone: "Asia/Manila" }),
            entries: (data.transaction.entries ?? []).map((e: any) => ({
                id: e.id, accountId: e.accountId, amount: Number(e.amount), type: e.type
            }))
        };
        editing = true;
    }

    // A two-entry transaction (most of them) has one amount: changing it changes both sides.
    const twoSided = $derived(draft.entries.length === 2);
    function setAmount(value: number) {
        for (const e of draft.entries) e.amount = value;
    }
    const draftBalanced = $derived(
        Math.abs(draft.entries.reduce((sum, e) => sum + (e.type === "DEBIT" ? 1 : -1) * (Number(e.amount) || 0), 0)) < 0.005
    );

    const accounts = $derived([...data.accounts].sort((a, b) => a.name.localeCompare(b.name)));

    function confirmDelete(event: SubmitEvent) {
        if (!confirm(`Delete "${data.transaction.description}"? This can't be undone.`)) {
            event.preventDefault();
        }
    }
</script>

<svelte:head>
    <title>{data.transaction.description} | FinanKal</title>
</svelte:head>

<div class="space-y-6">

    <!-- Header -->

    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">

        <div>

            <a
                    href={data.back}
                    class="text-sm text-slate-500 transition hover:text-slate-300"
            >
                ← {data.back === "/transactions" ? "Transactions" : "Account"}
            </a>

            <h1 class="mt-2 text-3xl font-bold tracking-tight text-white">
                {data.transaction.description}
            </h1>

            <p class="mt-1 text-sm text-slate-400">
                Transaction #{data.transaction.id}
            </p>

        </div>

        <div class="flex flex-wrap items-center gap-3">
            <span
                    class={`inline-flex rounded-full border px-3 py-1.5 text-sm font-medium ${
                    balanced
                        ? "border-emerald-500/20 bg-emerald-500/10 text-emerald-400"
                        : "border-red-500/20 bg-red-500/10 text-red-400"
                }`}
            >
                {balanced ? "✓ Balanced" : "✕ Not Balanced"}
            </span>

            {#if !editing}
                <button type="button" onclick={startEditing}
                        class="rounded-xl border border-slate-700 bg-slate-800 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-blue-500">
                    Edit
                </button>
            {/if}

            <form method="POST" action={`?/delete&from=${encodeURIComponent(data.back)}`} onsubmit={confirmDelete}>
                <button type="submit"
                        class="rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-2 text-sm font-medium text-red-300 transition hover:bg-red-500/20 focus-visible:outline-2 focus-visible:outline-red-500">
                    Delete
                </button>
            </form>
        </div>

    </div>

    {#if form?.error}
        <p role="alert" class="rounded-xl border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-300">
            {form.error}
        </p>
    {:else if form?.saved && !editing}
        <p role="status" class="text-sm text-emerald-400">Saved</p>
    {/if}


    {#if editing}

        <!-- Edit -->

        <form method="POST" action="?/update"
              class="space-y-5 rounded-2xl border border-slate-800 bg-slate-900 p-5"
              use:enhance={() => {
                  saving = true;
                  return async ({ result, update }) => {
                      saving = false;
                      if (result.type === "success") editing = false;
                      await update();
                  };
              }}>

            <input type="hidden" name="transaction" value={JSON.stringify(draft)} />

            <div class="grid gap-4 md:grid-cols-[1fr_12rem]">
                <label class="space-y-1.5">
                    <span class="text-sm text-slate-400">Description</span>
                    <input bind:value={draft.description} required
                           class="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-white focus:border-blue-500 focus:outline-none" />
                </label>
                <label class="space-y-1.5">
                    <span class="text-sm text-slate-400">Date</span>
                    <input type="date" bind:value={draft.date} required
                           class="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-white focus:border-blue-500 focus:outline-none" />
                </label>
            </div>

            {#if twoSided}
                <label class="block max-w-xs space-y-1.5">
                    <span class="text-sm text-slate-400">Amount</span>
                    <input type="number" min="0.01" step="0.01" required
                           value={draft.entries[0].amount}
                           oninput={(e) => setAmount(Number(e.currentTarget.value))}
                           class="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-right text-white tabular-nums focus:border-blue-500 focus:outline-none" />
                </label>
            {/if}

            <div class="space-y-3">
                {#each draft.entries as entry, i (entry.id)}
                    <div class="grid items-end gap-3 md:grid-cols-[6rem_1fr_10rem]">
                        <span class={`inline-flex w-fit rounded-full border px-2.5 py-1 text-xs font-medium ${typeClass(entry.type)}`}>
                            {entry.type}
                        </span>
                        <label class="space-y-1.5">
                            <span class="sr-only">Account for entry {i + 1}</span>
                            <select bind:value={entry.accountId}
                                    class="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-white focus:border-blue-500 focus:outline-none">
                                {#each accounts as account (account.id)}
                                    <option value={account.id}>{account.name}</option>
                                {/each}
                            </select>
                        </label>
                        {#if !twoSided}
                            <label class="space-y-1.5">
                                <span class="sr-only">Amount for entry {i + 1}</span>
                                <input type="number" min="0.01" step="0.01" required bind:value={entry.amount}
                                       class="w-full rounded-xl border border-slate-700 bg-slate-950 px-3 py-2 text-right text-white tabular-nums focus:border-blue-500 focus:outline-none" />
                            </label>
                        {/if}
                    </div>
                {/each}
            </div>

            {#if !draftBalanced}
                <p class="text-sm text-red-300">Debits and credits must add up to the same amount.</p>
            {/if}

            <div class="flex justify-end gap-3">
                <button type="button" onclick={() => (editing = false)}
                        class="rounded-xl border border-slate-700 px-4 py-2 text-sm font-medium text-slate-300 transition hover:bg-slate-800">
                    Cancel
                </button>
                <button type="submit" disabled={saving || !draftBalanced}
                        class="rounded-xl bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-500 disabled:opacity-50">
                    {saving ? "Saving…" : "Save"}
                </button>
            </div>
        </form>

    {:else}

    <!-- Transaction Information -->

    <div class="grid gap-4 md:grid-cols-3">

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

            <p class="text-sm text-slate-400">
                Date
            </p>

            <p class="mt-2 font-semibold text-white">
                {formatDate(data.transaction.createdAt)}
            </p>

        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

            <p class="text-sm text-slate-400">
                Total Debit
            </p>

            <p class="mt-2 text-2xl font-bold text-emerald-400">
                {currency.format(totalDebit)}
            </p>

        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

            <p class="text-sm text-slate-400">
                Total Credit
            </p>

            <p class="mt-2 text-2xl font-bold text-blue-400">
                {currency.format(totalCredit)}
            </p>

        </div>

    </div>


    <!-- Ledger Entries -->

    <div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">

        <div class="border-b border-slate-800 px-5 py-4">

            <h2 class="font-semibold text-white">
                Ledger Entries
            </h2>

            <p class="mt-1 text-sm text-slate-400">
                Accounts affected by this transaction.
            </p>

        </div>


        <div class="overflow-x-auto">

            <table class="w-full text-left">

                <thead class="border-b border-slate-800 bg-slate-950/50">

                <tr>

                    <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Account
                    </th>

                    <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Type
                    </th>

                    <th class="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Amount
                    </th>

                </tr>

                </thead>

                <tbody class="divide-y divide-slate-800">

                {#each data.transaction.entries ?? [] as entry}

                    <tr class="transition hover:bg-slate-800/40">

                        <td class="px-5 py-4">

                            <a
                                    href={`/accounts/${entry.accountId}`}
                                    class="font-medium text-white transition hover:text-blue-400"
                            >
                                {entry.accountName ?? entry.accountId}
                            </a>

                        </td>

                        <td class="px-5 py-4">

                                <span
                                        class={`inline-flex rounded-full border px-2.5 py-1 text-xs font-medium ${typeClass(entry.type)}`}
                                >
                                    {entry.type}
                                </span>

                        </td>

                        <td class="px-5 py-4 text-right font-semibold text-white">
                            {currency.format(entry.amount)}
                        </td>

                    </tr>

                {/each}

                </tbody>

            </table>

        </div>

    </div>

    {/if}


    <!-- Actions -->

    <div class="flex justify-between">

        <a
                href={data.back}
                class="rounded-xl border border-slate-700 bg-slate-900 px-5 py-3 font-medium text-slate-300 transition hover:bg-slate-800 hover:text-white"
        >
            ← Back to {data.back === "/transactions" ? "Transactions" : "Account"}
        </a>

    </div>

</div>
