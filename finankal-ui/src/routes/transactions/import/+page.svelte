<script lang="ts">
    import { enhance } from "$app/forms";
    import AccountBadge from "$lib/components/accounts/AccountBadge.svelte";
    import {
        FileText,
        Upload,
        X,
        LoaderCircle,
        CircleCheck,
        CircleAlert,
        TriangleAlert,
        Copy,
        Sparkles,
        Circle,
        CircleDashed
    } from "lucide-svelte";

    let { data, form } = $props();

    let bank = $state(""); // "" = detect from each statement
    let sourceAccountId = $state(""); // "" = find by the statement's account number
    let files = $state<File[]>([]);
    let fileError = $state("");
    let dragging = $state(false);
    let submitting = $state(false);
    let fileInput: HTMLInputElement;

    const canImport = $derived(files.length > 0 && !submitting);

    const imported = $derived(form?.results?.filter((r) => r.status === "IMPORTED") ?? []);
    const totalRecorded = $derived(imported.reduce((sum, r) => sum + r.imported, 0));

    function isPdf(f: File) {
        return f.type === "application/pdf" || f.name.toLowerCase().endsWith(".pdf");
    }

    /** Adds files to the list (same name and size = same file) and mirrors the list into the real input. */
    function add(list: FileList | null | undefined) {
        if (!list?.length) return;
        const chosen = Array.from(list);
        const notPdf = chosen.filter((f) => !isPdf(f));
        fileError = notPdf.length
            ? `${notPdf.map((f) => f.name).join(", ")} ${notPdf.length === 1 ? "isn't a PDF" : "aren't PDFs"}. Download statements as PDFs from your bank's app or website.`
            : "";
        const fresh = chosen.filter((f) => isPdf(f) && !files.some((g) => g.name === f.name && g.size === f.size));
        files = [...files, ...fresh];
        sync();
    }

    function remove(index: number) {
        files = files.filter((_, i) => i !== index);
        sync();
    }

    // The form submits whatever the input holds, so keep it equal to the list on screen.
    function sync() {
        const transfer = new DataTransfer();
        files.forEach((f) => transfer.items.add(f));
        fileInput.files = transfer.files;
    }

    function onDrop(event: DragEvent) {
        event.preventDefault();
        dragging = false;
        add(event.dataTransfer?.files);
    }

    function formatSize(bytes: number) {
        return bytes < 1024 * 1024
            ? `${Math.max(1, Math.round(bytes / 1024))} KB`
            : `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
    }

    const plural = (n: number, one: string, many: string) => `${n} ${n === 1 ? one : many}`;

    // "2026-10" → "Oct"
    const monthLabel = (month: string) =>
        new Date(`${month}-01T00:00:00Z`).toLocaleString("en", { month: "short", timeZone: "UTC" });
</script>

<svelte:head>
    <title>Import statements</title>
</svelte:head>

<div class="mx-auto max-w-3xl">

    <header class="mb-8">
        <h1 class="text-3xl font-bold tracking-tight text-white">
            Import statements
        </h1>
        <p class="mt-2 max-w-xl text-sm leading-6 text-slate-400">
            Upload the PDF statements from your banks and e-wallets, as many months as you like.
            FinanKal reads the bank and account number from each one and skips anything it already has.
        </p>
    </header>

    {#if form?.success}
        <section class="rounded-2xl border border-slate-800 bg-slate-900" aria-live="polite">

            <div class="flex items-start gap-4 p-6">
                <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-400">
                    <CircleCheck class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                    <h2 class="text-lg font-semibold text-white">
                        Recorded {plural(totalRecorded, "transaction", "transactions")}
                        from {plural(imported.length, "statement", "statements")}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-slate-400">
                        Each transaction is described the way it appears on the statement.
                    </p>
                </div>
            </div>

            <ul class="border-t border-slate-800">
                {#each form.results as result}
                    <li class="border-b border-slate-800 p-6 last:border-b-0">

                        <div class="flex items-start gap-4">
                            <div class={`flex h-12 w-10 shrink-0 items-center justify-center rounded-lg ${
                                result.status === "IMPORTED" ? "bg-emerald-500/10 text-emerald-400"
                                : result.status === "ALREADY_IMPORTED" ? "bg-slate-800 text-slate-400"
                                : "bg-red-500/10 text-red-400"
                            }`}>
                                {#if result.status === "IMPORTED"}
                                    <FileText class="h-5 w-5" />
                                {:else if result.status === "ALREADY_IMPORTED"}
                                    <Copy class="h-5 w-5" />
                                {:else}
                                    <CircleAlert class="h-5 w-5" />
                                {/if}
                            </div>
                            <div class="min-w-0 flex-1">
                                <p class="truncate text-sm font-semibold text-white">{result.fileName}</p>
                                <p class="mt-0.5 text-sm leading-6 text-slate-400">
                                    {#if result.status === "IMPORTED"}
                                        {plural(result.imported, "transaction", "transactions")} recorded in
                                        <span class="text-slate-200">{result.accountName}</span>.
                                        {#if !result.warnings.length}The balance matches the statement.{/if}
                                    {:else if result.status === "ALREADY_IMPORTED"}
                                        Already imported{result.bank ? ` (${result.bank})` : ""}, so it was skipped.
                                    {:else}
                                        Not imported{result.bank ? ` (${result.bank})` : ""}: {result.error}
                                    {/if}
                                </p>
                            </div>
                        </div>

                        {#if result.transfers.length}
                            <details class="mt-4 rounded-xl border border-blue-500/20 bg-blue-500/5">
                                <summary class="cursor-pointer px-4 py-2.5 text-sm font-medium text-slate-200">
                                    {plural(result.transfers.length, "transfer", "transfers")} between your accounts
                                    <span class="font-normal text-slate-500">· joined with the other bank's statement, not counted as spending or income</span>
                                </summary>
                                <ul class="divide-y divide-slate-800 border-t border-blue-500/20">
                                    {#each result.transfers as row}
                                        <li class="px-4 py-2.5 text-sm break-words text-slate-300 tabular-nums">{row}</li>
                                    {/each}
                                </ul>
                            </details>
                        {/if}

                        {#if result.duplicates.length}
                            <details class="mt-4 rounded-xl border border-slate-800 bg-slate-950/40">
                                <summary class="cursor-pointer px-4 py-2.5 text-sm font-medium text-slate-300">
                                    {plural(result.duplicates.length, "row was", "rows were")} already recorded
                                    <span class="font-normal text-slate-500">· same day, amount and direction</span>
                                </summary>
                                <ul class="divide-y divide-slate-800 border-t border-slate-800">
                                    {#each result.duplicates as row}
                                        <li class="px-4 py-2.5 text-sm break-words text-slate-400 tabular-nums">{row}</li>
                                    {/each}
                                </ul>
                            </details>
                        {/if}

                        {#if result.skipped.length}
                            <div class="mt-4">
                                <h3 class="flex items-center gap-2 text-sm font-semibold text-amber-400">
                                    <TriangleAlert class="h-4 w-4" />
                                    {plural(result.skipped.length, "row needs", "rows need")} your review
                                </h3>
                                <p class="mt-1 text-sm leading-6 text-slate-400">
                                    The statement doesn't show whether money went in or out on these, so they were left out.
                                </p>
                                <ul class="mt-3 divide-y divide-slate-800 overflow-hidden rounded-xl border border-amber-500/20 bg-amber-500/5">
                                    {#each result.skipped as row}
                                        <li class="px-4 py-2.5 text-sm break-words text-slate-300 tabular-nums">{row}</li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}

                        {#if result.warnings.length}
                            <div class="mt-4">
                                <h3 class="flex items-center gap-2 text-sm font-semibold text-amber-400">
                                    <CircleAlert class="h-4 w-4" />
                                    The balance doesn't match your bank
                                </h3>
                                <p class="mt-1 text-sm leading-6 text-slate-400">
                                    Import the missing month, or check transactions added by hand in that period.
                                </p>
                                <ul class="mt-3 divide-y divide-slate-800 overflow-hidden rounded-xl border border-amber-500/20 bg-amber-500/5">
                                    {#each result.warnings as warning}
                                        <li class="px-4 py-2.5 text-sm break-words text-slate-300 tabular-nums">{warning}</li>
                                    {/each}
                                </ul>
                            </div>
                        {/if}
                    </li>
                {/each}
            </ul>

            <div class="flex flex-col-reverse gap-3 border-t border-slate-800 p-6 sm:flex-row sm:justify-end">
                {#if form.results.some((r) => r.skipped.length)}
                    <a
                            href="/transactions/new"
                            class="inline-flex items-center justify-center rounded-xl border border-slate-700 bg-slate-800 px-4 py-2.5 text-sm font-semibold text-slate-200 transition hover:border-slate-600 hover:text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
                    >
                        Add a transaction
                    </a>
                {/if}
                <a
                        href="/transactions/import"
                        data-sveltekit-reload
                        class="inline-flex items-center justify-center rounded-xl border border-slate-700 bg-slate-800 px-4 py-2.5 text-sm font-semibold text-slate-200 transition hover:border-slate-600 hover:text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
                >
                    Import more statements
                </a>
                <a
                        href="/transactions"
                        class="inline-flex items-center justify-center rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500"
                >
                    View transactions
                </a>
            </div>

        </section>
    {:else}
        <form
                method="POST"
                action="?/import"
                enctype="multipart/form-data"
                class="rounded-2xl border border-slate-800 bg-slate-900"
                use:enhance={() => {
                    submitting = true;
                    return async ({ update }) => {
                        await update({ reset: false });
                        submitting = false;
                    };
                }}
        >

            <!-- Statement files -->
            <div class="p-6">
                <h2 class="text-lg font-semibold text-white" id="file-heading">Statement PDFs</h2>

                <input
                        bind:this={fileInput}
                        id="files"
                        name="files"
                        type="file"
                        multiple
                        accept="application/pdf,.pdf"
                        class="peer sr-only"
                        aria-labelledby="file-heading"
                        aria-describedby={fileError ? "file-error" : undefined}
                        onchange={(e) => add(e.currentTarget.files)}
                />

                <label
                        for="files"
                        class={`mt-4 flex cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed px-6 text-center transition peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-blue-500 ${
                            files.length ? "py-6" : "py-12"
                        } ${
                            dragging
                                ? "border-blue-500 bg-blue-600/10"
                                : "border-slate-700 bg-slate-950 hover:border-slate-600"
                        }`}
                        ondragenter={(e) => { e.preventDefault(); dragging = true; }}
                        ondragover={(e) => e.preventDefault()}
                        ondragleave={() => (dragging = false)}
                        ondrop={onDrop}
                >
                    <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-slate-800 text-slate-300">
                        <Upload class="h-5 w-5" />
                    </div>
                    <p class="mt-4 text-sm font-semibold text-white">
                        {files.length ? "Add more statements" : "Drop your statements here"}, or <span class="text-blue-400">browse files</span>
                    </p>
                    <p class="mt-1 text-xs text-slate-500">PDFs up to 10 MB each, any months, any banks</p>
                </label>

                {#if files.length}
                    <ul class="mt-4 divide-y divide-slate-800 overflow-hidden rounded-xl border border-slate-700 bg-slate-950" aria-label="Statements to import">
                        {#each files as f, i (f.name + f.size)}
                            <li class="flex items-center gap-4 p-3 pl-4">
                                <div class="flex h-10 w-8 shrink-0 items-center justify-center rounded-md bg-red-500/10 text-red-400">
                                    <FileText class="h-4 w-4" />
                                </div>
                                <div class="min-w-0 flex-1">
                                    <p class="truncate text-sm font-semibold text-white">{f.name}</p>
                                    <p class="mt-0.5 text-xs text-slate-400 tabular-nums">{formatSize(f.size)}</p>
                                </div>
                                <button
                                        type="button"
                                        onclick={() => remove(i)}
                                        class="flex h-9 w-9 items-center justify-center rounded-lg text-slate-400 transition hover:bg-slate-800 hover:text-white focus-visible:outline-2 focus-visible:outline-blue-500"
                                        aria-label={`Remove ${f.name}`}
                                >
                                    <X class="h-4 w-4" />
                                </button>
                            </li>
                        {/each}
                    </ul>
                {/if}

                {#if fileError}
                    <p id="file-error" class="mt-3 text-sm text-red-400">{fileError}</p>
                {/if}

                <div class="mt-5">
                    <label for="password" class="mb-2 block text-sm font-medium text-slate-300">
                        PDF password <span class="font-normal text-slate-500">(if they ask for one)</span>
                    </label>
                    <input
                            id="password"
                            name="password"
                            type="password"
                            autocomplete="off"
                            aria-describedby="password-hint"
                            class="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition placeholder:text-slate-600 focus:border-blue-500 sm:max-w-xs"
                    />
                    <p id="password-hint" class="mt-2 text-xs text-slate-500">
                        Some banks, like GCash and BDO, lock their statement PDFs. It's tried on every file and only used to open them.
                    </p>
                </div>
            </div>

            <!-- Bank -->
            <fieldset class="border-t border-slate-800 p-6">
                <legend class="text-lg font-semibold text-white">Bank</legend>
                <p class="mt-1 text-sm text-slate-400">Read from each statement. Choose one only if it's read wrong.</p>

                <div class="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-4">
                    {#each [{ value: "", label: "Detect" }, ...data.banks] as option}
                        <label
                                class={`flex h-12 cursor-pointer items-center justify-center gap-1.5 rounded-xl border px-2 text-sm font-semibold transition focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-blue-500 ${
                                bank === option.value
                                    ? "border-blue-500 bg-blue-600/15 text-white"
                                    : "border-slate-700 bg-slate-950 text-slate-300 hover:border-slate-600 hover:text-white"
                            }`}
                        >
                            <input
                                    type="radio"
                                    name="bank"
                                    value={option.value}
                                    bind:group={bank}
                                    class="sr-only"
                            />
                            {#if option.value === ""}<Sparkles class="h-3.5 w-3.5" />{/if}
                            {option.label}
                        </label>
                    {/each}
                </div>
            </fieldset>

            <!-- Account -->
            <div class="border-t border-slate-800 p-6">
                <label for="sourceAccountId" class="block text-lg font-semibold text-white">
                    Account
                </label>
                <p id="account-hint" class="mt-1 text-sm text-slate-400">
                    Matched by the account number on each statement. An account FinanKal hasn't seen yet is created for you.
                </p>

                <select
                        id="sourceAccountId"
                        name="sourceAccountId"
                        bind:value={sourceAccountId}
                        aria-describedby="account-hint account-routing"
                        class="mt-4 w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-white outline-none transition focus:border-blue-500 md:max-w-md"
                >
                    <option value="">Match by account number</option>
                    {#each data.sourceAccounts as account}
                        <option value={account.id}>All into {account.name}</option>
                    {/each}
                </select>

                <p id="account-routing" class="mt-3 text-xs leading-5 text-slate-500">
                    Choosing an account links it to the statement's account number, so next time it's matched automatically.
                    <!-- Names must match defaultExpenseAccount / defaultIncomeAccount in engine-go/internal/statement/service.go. -->
                    Money out is recorded against Outgoing Transfers, and money in against Incoming Transfers.
                </p>
            </div>

            {#if form?.error}
                <div role="alert" class="mx-6 mb-2 flex items-start gap-3 rounded-xl border border-red-500/20 bg-red-500/10 p-4 text-sm text-red-300">
                    <CircleAlert class="mt-0.5 h-4 w-4 shrink-0 text-red-400" />
                    {form.error}
                </div>
            {/if}

            <!-- Submit -->
            <div class="flex flex-col gap-4 border-t border-slate-800 p-6 sm:flex-row sm:items-center sm:justify-between">
                <p class="text-xs leading-5 text-slate-500 sm:max-w-sm">
                    Months can be in any order. Statements already imported, and transactions already recorded, are skipped.
                    A Maya Personal Loan disclosure statement adds the loan and its deadlines.
                </p>
                <button
                        type="submit"
                        disabled={!canImport}
                        class="inline-flex items-center justify-center gap-2 rounded-xl bg-blue-600 px-5 py-2.5 text-sm font-semibold text-white transition hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500 disabled:cursor-not-allowed disabled:bg-slate-800 disabled:text-slate-500"
                >
                    {#if submitting}
                        <LoaderCircle class="h-4 w-4 animate-spin motion-reduce:animate-none" />
                        Reading statements…
                    {:else}
                        Import {files.length > 1 ? `${files.length} statements` : "statement"}
                    {/if}
                </button>
            </div>

        </form>
    {/if}

    <!-- Which months each account has a statement for -->
    {#if data.checklist.rows.length}
        <section class="mt-8 rounded-2xl border border-slate-800 bg-slate-900" aria-labelledby="checklist-heading">
            <div class="p-6 pb-4">
                <h2 id="checklist-heading" class="text-lg font-semibold text-white">Statements on file</h2>
                <p class="mt-1 text-sm text-slate-400">
                    Months each account has a statement for. Fill the gaps so every transfer and balance can be checked.
                </p>
                <ul class="mt-3 flex flex-wrap gap-x-5 gap-y-1 text-xs text-slate-400">
                    <li class="flex items-center gap-1.5"><CircleCheck class="h-3.5 w-3.5 text-emerald-400" /> Whole month</li>
                    <li class="flex items-center gap-1.5"><CircleDashed class="h-3.5 w-3.5 text-amber-400" /> Part of the month</li>
                    <li class="flex items-center gap-1.5"><Circle class="h-3.5 w-3.5 text-slate-600" /> Missing</li>
                </ul>
                {#if data.checklist.rows.some((r) => r.accountType === "CREDIT_CARD")}
                    <p class="mt-2 text-xs text-slate-500">
                        Credit cards bill across months, so a card's month is ticked when the bill closing in it is on file.
                    </p>
                {/if}
            </div>

            <div class="overflow-x-auto border-t border-slate-800">
                <table class="w-full text-sm">
                    <thead>
                        <tr class="text-xs text-slate-500">
                            <th scope="col" class="sticky left-0 bg-slate-900 px-6 py-3 text-left font-medium">Account</th>
                            {#each data.checklist.months as month, i}
                                <th scope="col" class="px-2 py-3 text-center font-medium tabular-nums">
                                    {monthLabel(month)}
                                    {#if i === 0 || month.endsWith("-01")}
                                        <span class="block text-[10px] text-slate-600">{month.slice(0, 4)}</span>
                                    {/if}
                                </th>
                            {/each}
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-800">
                        {#each data.checklist.rows as row}
                            <tr>
                                <th scope="row" class="sticky left-0 bg-slate-900 px-6 py-3 text-left font-medium whitespace-nowrap text-slate-200">
                                    <span class="flex items-center gap-3">
                                        <AccountBadge name={row.accountName} type={row.accountType} />
                                        <span>
                                            {row.accountName}
                                            {#if row.bank}
                                                <span class="block text-xs font-normal text-slate-500">
                                                    {data.banks.find((b) => b.value === row.bank)?.label ?? row.bank}
                                                </span>
                                            {/if}
                                        </span>
                                    </span>
                                </th>
                                {#each row.cells as cell}
                                    {@const label = `${row.accountName}, ${monthLabel(cell.month)} ${cell.month.slice(0, 4)}: ${
                                        cell.coverage === "full"
                                            ? (row.accountType === "CREDIT_CARD" ? `bill on file (${cell.periods.join(", ")})` : "whole month")
                                        : cell.coverage === "partial" ? `part of the month (${cell.periods.join(", ")})`
                                        : "no statement"}`}
                                    <td class="px-2 py-3 text-center" title={label}>
                                        <span class="sr-only">{label}</span>
                                        {#if cell.coverage === "full"}
                                            <CircleCheck class="mx-auto h-5 w-5 text-emerald-400" aria-hidden="true" />
                                        {:else if cell.coverage === "partial"}
                                            <CircleDashed class="mx-auto h-5 w-5 text-amber-400" aria-hidden="true" />
                                        {:else}
                                            <Circle class="mx-auto h-5 w-5 text-slate-700" aria-hidden="true" />
                                        {/if}
                                    </td>
                                {/each}
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        </section>
    {/if}
</div>
