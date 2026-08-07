<script lang="ts">
    import TransactionEntryRow from "./TransactionEntryRow.svelte";

    interface Account {
        id: string;
        name: string;
        type: string;
        balance?: number;
    }

    interface Entry {
        accountId: string;
        amount: number;
        type: "DEBIT" | "CREDIT";
    }

    interface Props {
        accounts: Account[];
        entries: Entry[];
    }

    let {
        accounts,
        entries = $bindable()
    }: Props = $props();

    function addEntry() {
        entries = [
            ...entries,
            {
                accountId: "",
                amount: 0,
                type: "DEBIT"
            }
        ];
    }

    function removeEntry(index: number) {
        if (entries.length <= 2) return;

        entries = entries.filter((_, i) => i !== index);
    }

    const debitCount = $derived(
        entries.filter((e) => e.type === "DEBIT").length
    );

    const creditCount = $derived(
        entries.filter((e) => e.type === "CREDIT").length
    );
</script>

<div class="rounded-2xl border border-slate-800 bg-slate-900">

    <div class="flex items-center justify-between border-b border-slate-800 px-6 py-5">

        <div>

            <h2 class="text-lg font-semibold text-white">
                Ledger Entries
            </h2>

            <p class="mt-1 text-sm text-slate-400">
                Every transaction must have equal debit and credit totals.
            </p>

        </div>

        <button
                type="button"
                onclick={addEntry}
                class="rounded-xl bg-blue-600 px-4 py-2 text-sm font-medium text-white transition hover:bg-blue-500"
        >
            + Add Entry
        </button>

    </div>

    <div class="overflow-x-auto">

        <table class="min-w-full">

            <thead class="border-b border-slate-800 bg-slate-950/40">

            <tr class="text-left text-sm uppercase tracking-wide text-slate-400">

                <th class="px-5 py-3">
                    Account
                </th>

                <th class="w-44 px-5 py-3">
                    Type
                </th>

                <th class="w-52 px-5 py-3 text-right">
                    Amount
                </th>

                <th class="w-16 px-5 py-3"></th>

            </tr>

            </thead>

            <tbody>

            {#each entries as entry, index}

                <TransactionEntryRow
                        {entry}
                        {accounts}
                        onDelete={() => removeEntry(index)}
                        onUpdate={(updated) => {
                            entries[index] = updated;
                        }}
                />

            {/each}

            </tbody>

        </table>

    </div>

    <div class="flex items-center justify-between border-t border-slate-800 bg-slate-950/30 px-6 py-4">

        <div class="flex gap-6 text-sm">

            <div class="text-slate-400">
                Debit Entries
                <span class="ml-2 font-semibold text-emerald-400">
					{debitCount}
				</span>
            </div>

            <div class="text-slate-400">
                Credit Entries
                <span class="ml-2 font-semibold text-blue-400">
					{creditCount}
				</span>
            </div>

        </div>

        <div class="text-sm text-slate-500">
            Minimum of two ledger entries required.
        </div>

    </div>

</div>