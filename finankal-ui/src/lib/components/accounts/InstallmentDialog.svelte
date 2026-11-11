<script lang="ts">
    import { enhance } from "$app/forms";
    import { X } from "lucide-svelte";
    import { currency } from "$lib/utils/currency";
    import { formatDate } from "$lib/utils/date";

    interface Props {
        accountName: string;
        payFrom: { id: string; name: string }[];
        defaultPayFromId: string;
        defaultDisbursedAccountId: string;
        feeAccounts: { id: string; name: string }[];
        isLoan: boolean;
        disbursementAlreadyRecorded: boolean;
        installmentKind: "MAYA" | "UNIONBANK" | "FIXED";
        statementDay?: number;
        statementShift?: string;
        error?: string;
    }

    let { accountName, payFrom, defaultPayFromId, defaultDisbursedAccountId, feeAccounts, isLoan, disbursementAlreadyRecorded, installmentKind, statementDay, statementShift, error }: Props = $props();

    let dialog: HTMLDialogElement;
    let saving = $state(false);

    let sourceAccountId = $state(defaultPayFromId);
    let disbursedAccountId = $state(defaultDisbursedAccountId);
    let feeAccountId = $state(feeAccounts.find((a) => a.name === "Loan Fees")?.id ?? "");
    let recordDisbursement = $state(!disbursementAlreadyRecorded);
    let disbursedOn = $state("");
    let loanAmount = $state<number | null>(null);
    let disbursedAmount = $state<number | null>(null);
    let mode = $state<"fixed" | "lender">(installmentKind === "FIXED" ? "fixed" : "lender");
    let unionBankProduct = $state<"EasyConvert" | "EasyBill">("EasyBill");
    let monthlyAmount = $state<number | null>(null);
    let count = $state<number | null>(12);
    let firstDue = $state("");
    let mayaMonthlyEirPercent = $state<number | null>(null);
    let unionBankPrincipal = $state<number | null>(null);
    let unionBankFactorPercent = $state<number | null>(null);
    let processingFee = $state<number | null>(null);
    let unionBankAnnualEirPercent = $state<number | null>(null);
    let unionBankApprovalDate = $state("");
    let lenderError = $state("");
    let estimatedInterest = $state<number | null>(null);
    let unionBankMonthlyPayment = $state<number | null>(null);
    let unionBankMonthZeroInterest = $state<number | null>(null);
    let unionBankMonthZeroDays = $state(0);
    let unionBankFirstStatement = $state("");
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

    function statementDateIn(year: number, month: number): string {
        if (!(Number(statementDay) >= 1 && Number(statementDay) <= 31)) return "";
        const last = new Date(Date.UTC(year, month + 1, 0)).getUTCDate();
        const date = new Date(Date.UTC(year, month, Math.min(Number(statementDay), last)));
        if (statementShift === "NEXT_BUSINESS_DAY") {
            while (date.getUTCDay() === 0 || date.getUTCDay() === 6) date.setUTCDate(date.getUTCDate() + 1);
        } else if (statementShift === "NEAREST_BUSINESS_DAY") {
            if (date.getUTCDay() === 6) date.setUTCDate(date.getUTCDate() - 1);
            else if (date.getUTCDay() === 0) date.setUTCDate(date.getUTCDate() + 1);
        }
        return date.toISOString().slice(0, 10);
    }

    function nextStatementAfter(day: string): string {
        const [year, month] = day.split("-").map(Number);
        let date = statementDateIn(year, month - 1);
        if (date <= day) date = statementDateIn(year, month);
        return date;
    }

    /** Generates a fixed monthly amount, the lender terms used by UNO and MariBank. */
    function fill() {
        const n = Math.floor(Number(count));
        const cents = toCents(monthlyAmount);
        if (mode !== "fixed" || !firstDue || !(n >= 1 && n <= 360) || !(cents > 0)) { rows = []; return; }
        rows = Array.from({ length: n }, (_, i) => ({
            dueDate: addMonths(firstDue, i),
            amount: cents / 100
        }));
    }

    function calculateUnionBank() {
        rows = [];
        estimatedInterest = null;
        unionBankMonthlyPayment = null;
        unionBankMonthZeroInterest = null;
        unionBankMonthZeroDays = 0;
        unionBankFirstStatement = "";
        const principalCents = toCents(unionBankPrincipal);
        const feeCents = toCents(processingFee);
        const n = Math.floor(Number(count));
        if (!unionBankApprovalDate || !(Number(statementDay) >= 1 && Number(statementDay) <= 31) || !(principalCents > 0) ||
            !(unionBankAnnualEirPercent != null && Number.isFinite(unionBankAnnualEirPercent) && unionBankAnnualEirPercent >= 0) ||
            !(unionBankFactorPercent != null && Number.isFinite(unionBankFactorPercent) && unionBankFactorPercent >= 0) ||
            !(processingFee != null && Number.isFinite(processingFee) && feeCents >= 0) ||
            !(n >= 1 && n <= 360)) {
            lenderError = "Enter the balance, annual EIR, converted monthly factor rate, term, processing fee and approval date; set the card statement day.";
            return;
        }
        const firstStatement = nextStatementAfter(unionBankApprovalDate);
        if (!firstStatement) { lenderError = "Set a valid statement day on this card."; return; }
        const monthZeroDays = (Date.parse(`${firstStatement}T00:00:00Z`) - Date.parse(`${unionBankApprovalDate}T00:00:00Z`)) / 86400000;
        if (!(monthZeroDays > 0)) { lenderError = "The first statement must be after the approval date."; return; }
        const monthlyEir = Math.pow(1 + unionBankAnnualEirPercent / 100, 1 / 12) - 1;
        const monthZeroInterestCents = Math.round(principalCents * monthlyEir * monthZeroDays / 30);
        const installmentCents = Math.round(principalCents * (1 / n + unionBankFactorPercent / 100));
        const installmentInterestCents = installmentCents * n - principalCents;
        if (installmentInterestCents < 0) { lenderError = "This rate produces a total repayment below the principal."; return; }
        const firstStatementCents = feeCents + monthZeroInterestCents;
        rows = [
            ...(firstStatementCents > 0 ? [{ dueDate: firstStatement, amount: firstStatementCents / 100 }] : []),
            ...Array.from({ length: n }, (_, i) => {
                const [year, month] = firstStatement.split("-").map(Number);
                return { dueDate: statementDateIn(year, month - 1 + i + 1), amount: installmentCents / 100 };
            })
        ];
        estimatedInterest = (installmentInterestCents + monthZeroInterestCents) / 100;
        unionBankMonthlyPayment = installmentCents / 100;
        unionBankMonthZeroInterest = monthZeroInterestCents / 100;
        unionBankMonthZeroDays = monthZeroDays;
        unionBankFirstStatement = firstStatement;
        customName = `${accountName} ${unionBankProduct}`;
        lenderError = "";
    }

    function calculateMaya() {
        rows = [];
        estimatedInterest = null;
        const principalCents = toCents(loanAmount);
        const n = Math.floor(Number(count));
        if (!firstDue || !disbursedOn || !(principalCents > 0) || !(mayaMonthlyEirPercent != null && Number.isFinite(mayaMonthlyEirPercent) && mayaMonthlyEirPercent >= 0) || !(n >= 1 && n <= 360)) {
            lenderError = "Enter the loan principal, monthly EIR, term, disbursement date and first due date.";
            return;
        }
        const dailyRate = Math.pow(1 + mayaMonthlyEirPercent / 100, 12 / 365) - 1;
        const dates = Array.from({ length: n }, (_, i) => addMonths(firstDue, i));
        if (dates[0] <= disbursedOn) {
            lenderError = "The first due date must be after the disbursement date.";
            return;
        }
        const dayCounts = dates.map((date, i) => {
            const start = i === 0 ? disbursedOn : dates[i - 1];
            return (Date.parse(`${date}T00:00:00Z`) - Date.parse(`${start}T00:00:00Z`)) / 86400000;
        });
        const remainingAfter = (paymentCents: number) => {
            let balance = principalCents;
            for (const days of dayCounts) {
                const interest = Math.round(balance * dailyRate * days);
                balance = Math.max(0, balance + interest - paymentCents);
            }
            return balance;
        };
        let low = Math.ceil(principalCents / n);
        let high = Math.ceil(principalCents * (1 + dailyRate * 400) / n);
        while (remainingAfter(high) > 0) high *= 2;
        while (low < high) {
            const mid = Math.floor((low + high) / 2);
            if (remainingAfter(mid) <= 0) high = mid;
            else low = mid + 1;
        }
        let balance = principalCents;
        rows = dates.map((dueDate, i) => {
            const interest = Math.round(balance * dailyRate * dayCounts[i]);
            const amountCents = i === n - 1 ? balance + interest : low;
            balance = Math.max(0, balance + interest - amountCents);
            return { dueDate, amount: amountCents / 100 };
        });
        estimatedInterest = rows.reduce((total, row) => total + toCents(row.amount), 0) / 100 - principalCents / 100;
        customName = null;
        lenderError = "";
    }

    function invalidateCalculation() {
        rows = [];
        estimatedInterest = null;
        unionBankMonthlyPayment = null;
        unionBankMonthZeroInterest = null;
        unionBankMonthZeroDays = 0;
        unionBankFirstStatement = "";
        customName = null;
        lenderError = "";
    }

    function disbursementToggled() {
        if (installmentKind === "MAYA" && mode === "lender") invalidateCalculation();
    }

    function loanAmountChanged() {
        if (installmentKind === "MAYA" && mode === "lender") invalidateCalculation();
    }

    function setMode(next: "fixed" | "lender") {
        mode = installmentKind === "FIXED" ? "fixed" : next;
        rows = [];
        customName = null;
        estimatedInterest = null;
        unionBankMonthlyPayment = null;
        lenderError = "";
    }

    const sum = $derived(rows.reduce((cents, r) => cents + (toCents(r.amount) || 0), 0) / 100);
    const loanFees = $derived((toCents(loanAmount) - toCents(disbursedAmount)) / 100);
    const suggestedName = $derived(`${accountName} ${currency.format(sum).replace(".00", "")} (${rows.length} ${rows.length === 1 ? "installment" : "installments"}${rows[0]?.dueDate ? `, from ${formatDate(rows[0].dueDate)}` : ""})`);
    const name = $derived(customName ?? suggestedName);

    // The first thing stopping the schedule from being saved, in the order the form reads.
    const problem = $derived.by(() => {
        if (!sourceAccountId) return "Choose the account you pay from.";
        if (isLoan && recordDisbursement) {
            if (!(toCents(loanAmount) > 0)) return "Enter the full loan amount.";
            if (!(toCents(disbursedAmount) > 0) || toCents(disbursedAmount) > toCents(loanAmount)) return "Disbursed amount must be above zero and no more than the loan amount.";
            if (!disbursedAccountId) return "Choose where the loan was disbursed.";
            if (!disbursedOn) return "Enter the disbursement date.";
            if (loanFees > 0 && !feeAccounts.length) return "Add an expense account before recording deducted loan fees.";
            if (loanFees > 0 && !feeAccountId) return "Choose an expense category for the deducted fees.";
        }
        if (isLoan && !disbursementAlreadyRecorded && !recordDisbursement) return "Record the disbursement for this new loan.";
        if (!rows.length) return mode === "fixed" ? "Enter a monthly payment, installment count and first due date." : "Add the lender's monthly payment schedule.";
        for (const [i, r] of rows.entries()) {
            if (!r.dueDate) return `Installment ${i + 1} needs a due date.`;
            if (!(toCents(r.amount) > 0)) return `Installment ${i + 1} needs an amount above zero.`;
            if (i > 0 && r.dueDate <= rows[i - 1].dueDate) return `Installment ${i + 1} must be due after installment ${i}.`;
        }
        if (!name.trim()) return "Give the schedule a name.";
        return null;
    });

    const payload = $derived(JSON.stringify({
        name: name.trim(), sourceAccountId, installments: rows,
        ...(isLoan && recordDisbursement ? { disbursement: { loanAmount, disbursedAmount, accountId: disbursedAccountId, date: disbursedOn, feeAccountId } } : {})
    }));

    function reset() {
        mode = installmentKind === "FIXED" ? "fixed" : "lender";
        unionBankProduct = "EasyBill";
        sourceAccountId = defaultPayFromId;
        disbursedAccountId = defaultDisbursedAccountId;
        recordDisbursement = !disbursementAlreadyRecorded;
        feeAccountId = feeAccounts.find((a) => a.name === "Loan Fees")?.id ?? "";
        disbursedOn = "";
        loanAmount = null;
        disbursedAmount = null;
        monthlyAmount = null;
        count = 12;
        firstDue = "";
        mayaMonthlyEirPercent = null;
        unionBankPrincipal = null;
        unionBankFactorPercent = null;
        processingFee = null;
        unionBankAnnualEirPercent = null;
        unionBankApprovalDate = "";
        lenderError = "";
        estimatedInterest = null;
        unionBankMonthlyPayment = null;
        unionBankMonthZeroInterest = null;
        unionBankMonthZeroDays = 0;
        unionBankFirstStatement = "";
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

            {#if isLoan}
                <fieldset class="space-y-3 rounded-xl border border-slate-800 p-4">
                    <legend class="px-1 text-sm font-medium text-slate-300">Loan disbursement</legend>
                    {#if disbursementAlreadyRecorded}
                        <p class="rounded-lg bg-emerald-950/40 px-3 py-2 text-xs text-emerald-300">A previous disbursement is recorded. Leave this unchecked to add a schedule only, or check it to record a new loan disbursement.</p>
                    {/if}
                    <label class="flex items-center gap-2 text-sm text-slate-300"><input type="checkbox" bind:checked={recordDisbursement} onchange={disbursementToggled} class="accent-blue-500" />Record a new loan disbursement transaction</label>
                    <div class="grid gap-3 sm:grid-cols-2">
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">Loan amount / starting principal (₱)</span><input type="number" min="0.01" step="0.01" bind:value={loanAmount} oninput={loanAmountChanged} class={field} /></label>
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">Amount disbursed to you (₱)</span><input type="number" min="0.01" step="0.01" bind:value={disbursedAmount} disabled={!recordDisbursement} class={field} /></label>
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">Disbursed to</span><select bind:value={disbursedAccountId} class={field} required={recordDisbursement} disabled={!recordDisbursement}><option value="" disabled>Choose a bank or wallet</option>{#each payFrom as a (a.id)}<option value={a.id}>{a.name}</option>{/each}</select></label>
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">Disbursement date</span><input type="date" bind:value={disbursedOn} disabled={!recordDisbursement} class={field} /></label>
                        {#if recordDisbursement && loanFees > 0 && feeAccounts.length}
                            <label class="block sm:col-span-2"><span class="mb-1 block text-xs text-slate-400">Deducted fees: {currency.format(loanFees)} · record as</span><select bind:value={feeAccountId} class={field} required><option value="" disabled>Choose an expense category</option>{#each feeAccounts as a (a.id)}<option value={a.id}>{a.name}</option>{/each}</select></label>
                        {:else if recordDisbursement && loanFees > 0}
                            <p role="alert" class="text-xs text-amber-300 sm:col-span-2">No expense accounts are available to record the deducted fees.</p>
                        {/if}
                    </div>
                    <p class="text-xs text-slate-500">{recordDisbursement ? "The principal, amount received and deducted fees will be recorded as one ledger transaction." : "No new disbursement transaction will be recorded with this schedule."}</p>
                </fieldset>
            {/if}

            {#if installmentKind !== "FIXED"}
                <div class="grid grid-cols-2 gap-2" role="group" aria-label="Installment entry method">
                    <button type="button" aria-pressed={mode === "fixed"} onclick={() => setMode("fixed")}
                            class={`rounded-xl border px-3 py-2 text-sm font-medium ${mode === "fixed" ? "border-blue-500 bg-blue-500/10 text-blue-300" : "border-slate-700 text-slate-400"}`}>
                        Fixed monthly
                    </button>
                    <button type="button" aria-pressed={mode === "lender"} onclick={() => setMode("lender")}
                            class={`rounded-xl border px-3 py-2 text-sm font-medium ${mode === "lender" ? "border-blue-500 bg-blue-500/10 text-blue-300" : "border-slate-700 text-slate-400"}`}>
                        Lender schedule
                    </button>
                </div>
            {/if}

            {#if mode === "fixed"}
                <fieldset>
                    <legend class="mb-1.5 text-sm font-medium text-slate-300">For UNO Digital Bank, MariBank and SPayLater fixed payments</legend>
                    <div class="grid gap-3 sm:grid-cols-3">
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">Monthly payment (₱)</span>
                            <input type="number" inputmode="decimal" min="0" step="0.01" placeholder="3,723.50" bind:value={monthlyAmount} oninput={fill} class={field} />
                        </label>
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">Number of payments</span>
                            <input type="number" inputmode="numeric" min="1" max="360" step="1" bind:value={count} oninput={fill} class={field} />
                        </label>
                        <label class="block"><span class="mb-1 block text-xs text-slate-400">First due date</span>
                            <input type="date" bind:value={firstDue} oninput={fill} class={field} />
                        </label>
                    </div>
                    <p class="mt-2 text-xs text-slate-500">Repeats the lender's fixed payment monthly. Total is calculated from payment × count.</p>
                </fieldset>
            {:else}
                <fieldset>
                    <legend class="mb-1.5 text-sm font-medium text-slate-300">Choose the lender schedule</legend>
                    <p class="rounded-lg bg-slate-950/50 px-3 py-2 text-sm font-medium text-slate-200">{installmentKind === "MAYA" ? "Maya Personal Loan" : "UnionBank Credit Card Installment"}</p>
                    {#if installmentKind === "MAYA"}
                        <p class="mt-2 text-xs text-slate-400">The monthly EIR is converted to an equivalent daily rate and accrues on the remaining principal for the actual days between disbursement and due dates. The estimate solves for level payments; the final payment may differ by a few centavos. Fees and overdue penalties are excluded.</p>
                        <div class="mt-3 grid gap-3 sm:grid-cols-3">
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Monthly effective interest rate (EIR, %)</span><input type="number" min="0" step="0.01" placeholder="1.40" bind:value={mayaMonthlyEirPercent} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Number of payments</span><input type="number" min="1" max="360" step="1" bind:value={count} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">First due date</span><input type="date" bind:value={firstDue} oninput={invalidateCalculation} class={field} /></label>
                        </div>
                        <button type="button" onclick={calculateMaya} class="mt-3 rounded-lg bg-slate-800 px-3 py-2 text-sm text-blue-300 hover:bg-slate-700">Calculate Maya schedule</button>
                        {#if estimatedInterest !== null}<p class="mt-2 text-xs text-slate-300">Estimated monthly payment: {currency.format(rows[0]?.amount ?? 0)} · Total interest: {currency.format(estimatedInterest)}</p>{/if}
                    {:else}
                        <p class="mt-2 text-xs text-slate-400">The monthly installment uses principal × [1 ÷ term + CMFR]. Month 0 interest uses principal × monthly EIR × days to the first statement ÷ 30. Regular installments start on the following statement.</p>
                        <div class="mt-3 grid gap-3 sm:grid-cols-2">
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Plan type</span><select bind:value={unionBankProduct} class={field} onchange={invalidateCalculation}><option value="EasyConvert">EasyConvert · transaction</option><option value="EasyBill">EasyBill · statement balance</option></select></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">{unionBankProduct === "EasyBill" ? "Statement balance (₱)" : "Transaction principal (₱)"}</span><input type="number" min="0" step="0.01" bind:value={unionBankPrincipal} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Converted monthly factor rate (%)</span><input type="number" min="0" step="0.01" placeholder="1.00" bind:value={unionBankFactorPercent} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Term (months)</span><input type="number" min="1" max="360" step="1" bind:value={count} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Annual EIR (% p.a.)</span><input type="number" min="0" step="0.01" placeholder="33.23" bind:value={unionBankAnnualEirPercent} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Processing fee (₱, enter 0 if waived)</span><input type="number" min="0" step="0.01" bind:value={processingFee} oninput={invalidateCalculation} class={field} /></label>
                            <label class="block"><span class="mb-1 block text-xs text-slate-400">Approval date</span><input type="date" bind:value={unionBankApprovalDate} oninput={invalidateCalculation} class={field} /></label>
                        </div>
                        <button type="button" onclick={calculateUnionBank} class="mt-3 rounded-lg bg-slate-800 px-3 py-2 text-sm text-blue-300 hover:bg-slate-700">Calculate UnionBank schedule</button>
                        {#if unionBankMonthlyPayment !== null}<p class="mt-2 text-xs text-slate-300">First statement: {formatDate(unionBankFirstStatement)} ({unionBankMonthZeroDays} days after approval) · Month 0 interest: {currency.format(unionBankMonthZeroInterest ?? 0)} · First statement charge: {currency.format((Number(processingFee) || 0) + (unionBankMonthZeroInterest ?? 0))} · Monthly installment: {currency.format(unionBankMonthlyPayment)} · Total interest: {currency.format(estimatedInterest ?? 0)}</p>{/if}
                        {#if !(Number(statementDay) >= 1 && Number(statementDay) <= 31)}<p class="mt-2 text-xs text-amber-300">Set this card's statement day to calculate the first statement date.</p>{/if}
                    {/if}
                    {#if lenderError}<p role="alert" class="mt-2 text-xs text-amber-300">{lenderError}</p>{/if}
                </fieldset>
            {/if}

            <section aria-label="Schedule">
                {#if rows.length}
                    <div class="grid grid-cols-[2rem_1fr_1fr_2rem] items-center gap-x-2 border-b border-slate-800 pb-2 text-xs text-slate-500">
                        <span>{installmentKind === "UNIONBANK" ? "Month" : "#"}</span><span>Due</span><span>Amount (₱)</span><span></span>
                    </div>
                    <ol class="divide-y divide-slate-800/70">
                        {#each rows as row, i}
                            <li class="grid grid-cols-[2rem_1fr_1fr_2rem] items-center gap-x-2 py-1.5">
                                <span class="text-sm text-slate-500 tabular-nums">{installmentKind === "UNIONBANK" && unionBankFirstStatement ? (Number(row.dueDate.slice(0, 4)) - Number(unionBankFirstStatement.slice(0, 4))) * 12 + Number(row.dueDate.slice(5, 7)) - Number(unionBankFirstStatement.slice(5, 7)) : i + 1}</span>
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
