<script lang="ts">
    let {
        description = $bindable(""),
        transactionDate = $bindable("")
    } = $props();
</script>

<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">

    <div class="mb-6">
        <h2 class="text-lg font-semibold text-white">
            Transaction Details
        </h2>

        <p class="mt-1 text-sm text-slate-400">
            Enter the basic information for this transaction.
        </p>
    </div>

    <div class="grid gap-5 md:grid-cols-2">

        <div class="md:col-span-2">
            <label
                    for="description"
                    class="mb-2 block text-sm font-medium text-slate-300"
            >
                Description
            </label>

            <input
                    id="description"
                    name="description"
                    type="text"
                    bind:value={description}
                    placeholder="e.g. Grocery shopping"
                    class="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition placeholder:text-slate-600 focus:border-blue-500"
                    required
            />
        </div>

        <div>
            <label
                    for="transactionDate"
                    class="mb-2 block text-sm font-medium text-slate-300"
            >
                Transaction Date
            </label>

            <input
                    id="transactionDate"
                    name="transactionDate"
                    type="date"
                    bind:value={transactionDate}
                    class="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition focus:border-blue-500"
            />
        </div>

    </div>

</div>