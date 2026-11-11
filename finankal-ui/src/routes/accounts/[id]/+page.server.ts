import { error, fail, redirect } from "@sveltejs/kit";
import { deleteLoan, getAccount, getLedgerEntries, getLoanDetails, getUserAccounts, updateLoanDetails } from "$lib/server/api/accounts";
import { createInstallments, deleteInstallments, getPayables, payCardInstallment, updateInstallments } from "$lib/server/api/payables";
import { createTransaction, deleteTransaction, getTransaction } from "$lib/server/api/transactions";
import type { Actions, PageServerLoad } from "./$types";

const USER_ID = "bcc539f6-fc94-4a2e-9ce6-3c03c0560204";

const isDisbursementFor = (description: string | undefined, accountName: string) => {
    const text = description?.trim().toLowerCase() ?? "";
    return text.startsWith(`${accountName.trim().toLowerCase()} `) && text.endsWith(" disbursement");
};

export const load: PageServerLoad = async ({ params }) => {
    try {
        const [account, entries, payables, accounts] = await Promise.all([
            getAccount(params.id),
            getLedgerEntries(params.id),
            // The page still loads without deadlines if they can't be fetched.
            getPayables(USER_ID).catch((e) => {
                console.error("Payables failed to load:", e);
                return [];
            }),
            // Only needed to pick where installments are paid from.
            getUserAccounts(USER_ID).catch((e) => {
                console.error("Accounts failed to load:", e);
                return [];
            })
        ]);
        const loan = account.type === "LIABILITY" ? await getLoanDetails(USER_ID, params.id).catch(() => null) : null;
        const assets = accounts.filter((a) => a.type === "ASSET");
        const linkedEntry = (matches: (description: string | undefined) => boolean) => entries.find((entry) => matches(entry.description) && entry.transactionId);
        const disbursementEntry = account.type === "LIABILITY" ? linkedEntry((description) => isDisbursementFor(description, account.name)) : undefined;
        const repaymentEntry = account.type === "LIABILITY" ? linkedEntry((description) => description?.toLowerCase().includes("loan installment payment") ?? false) : undefined;
        const [disbursementTransaction, repaymentTransaction] = await Promise.all([
            disbursementEntry?.transactionId ? getTransaction(disbursementEntry.transactionId).catch(() => null) : null,
            repaymentEntry?.transactionId ? getTransaction(repaymentEntry.transactionId).catch(() => null) : null
        ]);
        const assetCounterpart = (transaction: typeof disbursementTransaction) =>
            transaction?.entries.find((entry) => assets.some((asset) => asset.id === entry.accountId))?.accountId;
        const scheduledRepaymentAccount = payables.find((payable) => payable.accountId === params.id && payable.sourceAccountId && assets.some((asset) => asset.id === payable.sourceAccountId))?.sourceAccountId;
        const commonRepaymentAccount = payables.find((payable) => payable.sourceAccountId && assets.some((asset) => asset.id === payable.sourceAccountId))?.sourceAccountId;
        const soleAssetAccount = assets.length === 1 ? assets[0].id : "";
        const accountKey = account.name.toLowerCase().replace(/[^a-z0-9]/g, "");
        const equivalentName = account.type === "CREDIT_CARD" && accountKey.includes("unionbank") ? "unionbankpayroll"
            : accountKey.includes("maya") ? "mayabank"
            : accountKey.includes("uno") ? "gcash"
            : accountKey.includes("maribank") || accountKey.includes("mariloan") ? "maribank"
            : "";
        const equivalentAccountId = equivalentName ? assets.find((asset) => asset.name.toLowerCase().replace(/[^a-z0-9]/g, "").includes(equivalentName))?.id : undefined;
        const defaultPayFromId = equivalentAccountId ?? scheduledRepaymentAccount ?? assetCounterpart(repaymentTransaction) ?? commonRepaymentAccount ?? soleAssetAccount;
        const defaultDisbursedAccountId = equivalentAccountId ?? assetCounterpart(disbursementTransaction) ?? assetCounterpart(repaymentTransaction) ?? defaultPayFromId;
        const installmentKind = account.type === "CREDIT_CARD" && accountKey.includes("unionbank") ? "UNIONBANK"
            : account.type === "LIABILITY" && accountKey.includes("maya") ? "MAYA"
            : "FIXED";
        const card = accounts.find((account) => account.id === params.id && account.type === "CREDIT_CARD");

        return {
            account: card ? { ...account, creditLimit: card.creditLimit, billingDay: card.billingDay, statementShift: card.statementShift } : account,
            entries,
            loan,
            disbursementAlreadyRecorded: account.type === "LIABILITY" && entries.some((entry) => isDisbursementFor(entry.description, account.name)),
            // This account's own deadlines (a loan's installments, a card's bills), soonest first.
            deadlines: payables.filter((p) => p.accountId === params.id),
            // Installments are paid from a bank or wallet.
            payFrom: accounts.filter((a) => a.type === "ASSET").map(({ id, name }) => ({ id, name })),
            feeAccounts: accounts.filter((a) => a.type === "EXPENSE").map(({ id, name }) => ({ id, name })),
            defaultPayFromId,
            defaultDisbursedAccountId,
            installmentKind
        };
    } catch (e) {
        console.error(e);
        throw error(404, "Account not found");
    }
};

export const actions: Actions = {
    payInstallment: async ({ request, params }) => {
        const fields = await request.formData();
        const name = fields.get("name");
        const dueDate = fields.get("dueDate");
        const installmentNumber = fields.get("installmentNumber");
        const sourceAccountId = fields.get("sourceAccountId");
        const scheduleId = fields.get("scheduleId");
        const occurrenceId = fields.get("occurrenceId");
        const paymentAmount = fields.get("paymentAmount");
        if (typeof name !== "string" || typeof dueDate !== "string" ||
            typeof installmentNumber !== "string" || typeof sourceAccountId !== "string" || typeof scheduleId !== "string" || typeof occurrenceId !== "string") {
            return fail(400, { payError: "This installment is missing payment details. Refresh and try again." });
        }

        try {
            // Refresh Maya's accrued interest before using its ledger balance as the overpayment cap.
            const payables = await getPayables(USER_ID);
            const account = await getAccount(params.id);
            if (account.type !== "LIABILITY" && account.type !== "CREDIT_CARD") return fail(400, { payError: "Only loan or card installments can be paid here." });
            const oldest = payables.find((payable) => payable.accountId === params.id && payable.scheduleId === scheduleId);
            if (!oldest || oldest.name !== name || oldest.dueDate !== dueDate ||
                String(oldest.installmentNumber ?? "") !== installmentNumber ||
                oldest.sourceAccountId !== sourceAccountId || oldest.occurrenceId !== occurrenceId || !oldest.sourceAccountId || !oldest.occurrenceId) {
                return fail(409, { payError: "This is no longer the next unpaid installment in this plan. Refresh and try again." });
            }

            let amount = Number(oldest.amount);
            if (!(amount > 0 && Number.isFinite(amount))) return fail(400, { payError: "This installment has an invalid amount." });
            if (paymentAmount !== null) {
                if (typeof paymentAmount !== "string" || !Number.isFinite(Number(paymentAmount))) return fail(400, { payError: "Enter a valid payment amount." });
                const requestedCents = Math.round(Number(paymentAmount) * 100);
                const dueCents = Math.round(amount * 100);
                const mayaLoan = account.type === "LIABILITY" && account.name.toLowerCase().includes("maya")
                    ? await getLoanDetails(USER_ID, params.id).catch(() => null)
                    : null;
                if (mayaLoan) {
                    const outstandingCents = Math.round(Math.max(0, -account.balance) * 100);
                    if (requestedCents < dueCents) return fail(400, { payError: "Payment must cover at least the next installment." });
                    if (requestedCents > outstandingCents) return fail(400, { payError: "Payment cannot exceed the current outstanding loan balance." });
                    amount = requestedCents / 100;
                } else if (requestedCents !== dueCents) {
                    return fail(400, { payError: "Only a registered Maya Personal Loan can be paid ahead." });
                }
            }
            const number = oldest.installmentNumber ? ` #${oldest.installmentNumber}` : "";
            const today = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10);
            if (account.type === "CREDIT_CARD") await payCardInstallment(USER_ID, oldest.occurrenceId);
            else await createTransaction({
                    description: `Loan installment payment${number}${amount > Number(oldest.amount) ? " (advance)" : ""}`, date: today,
                    entries: [{ accountId: oldest.accountId, amount, type: "DEBIT" }, { accountId: oldest.sourceAccountId, amount, type: "CREDIT" }]
                });
            return { paid: true };
        } catch (e) {
            console.error("Paying loan installment failed:", e);
            return fail(400, { payError: reason(e, "Couldn't record the payment. Check that the API and engine are running.") });
        }
    },
    addInstallments: async ({ request, params }) => {
        const raw = (await request.formData()).get("schedule");
        if (typeof raw !== "string") {
            return fail(400, { error: "The schedule didn't reach the server. Try again." });
        }

        let schedule;
        try {
            schedule = JSON.parse(raw);
        } catch {
            return fail(400, { error: "The schedule didn't reach the server. Try again." });
        }

        try {
            const { disbursement, ...installmentSchedule } = schedule;
            const account = await getAccount(params.id);
            let transactionId: string | undefined;
            if (account.type === "LIABILITY") {
                const existingEntries = await getLedgerEntries(params.id);
                const alreadyRecorded = existingEntries.some((entry) => isDisbursementFor(entry.description, account.name));
                if (!alreadyRecorded && disbursement == null) return fail(400, { error: "Enter the loan disbursement details before adding its installments." });
                if (disbursement != null) {
                    const principalInput = Number(disbursement.loanAmount);
                    const receivedInput = Number(disbursement.disbursedAmount);
                    const principalCents = Math.round(principalInput * 100);
                    const receivedCents = Math.round(receivedInput * 100);
                    const principal = principalCents / 100;
                    const received = receivedCents / 100;
                    const fees = (principalCents - receivedCents) / 100;
                    const date = disbursement.date;
                    if (!Number.isFinite(principalInput) || !Number.isFinite(receivedInput) || principal <= 0 || received <= 0 || received > principal ||
                        Math.abs(principalInput * 100 - principalCents) > 0.0001 || Math.abs(receivedInput * 100 - receivedCents) > 0.0001 ||
                        typeof date !== "string" || !/^\d{4}-\d{2}-\d{2}$/.test(date) || typeof disbursement.accountId !== "string") {
                        return fail(400, { error: "Enter valid loan and disbursed amounts, destination account and disbursement date." });
                    }
                    const accounts = await getUserAccounts(USER_ID);
                    if (!accounts.some((a) => a.id === disbursement.accountId && a.type === "ASSET")) {
                        return fail(400, { error: "Choose one of your bank or wallet accounts for the disbursement." });
                    }
                    if (fees > 0 && !accounts.some((a) => a.id === disbursement.feeAccountId && a.type === "EXPENSE")) {
                        return fail(400, { error: "Choose an expense account for the fees deducted from the loan." });
                    }
                    const entries = [
                        { accountId: disbursement.accountId, amount: received, type: "DEBIT" as const },
                        { accountId: params.id, amount: principal, type: "CREDIT" as const },
                        ...(fees > 0 ? [{ accountId: disbursement.feeAccountId, amount: fees, type: "DEBIT" as const }] : [])
                    ];
                    const transaction = await createTransaction({
                        description: `${String(schedule.name).trim()} disbursement`, date, entries
                    });
                    transactionId = transaction.id;
                }
            } else if (disbursement != null) {
                return fail(400, { error: "Only a loan account can have a loan disbursement." });
            }
            try {
                await createInstallments(USER_ID, { ...installmentSchedule, accountId: params.id });
            } catch (e) {
                if (transactionId) {
                    try { await deleteTransaction(USER_ID, transactionId); }
                    catch { return fail(500, { error: "The disbursement transaction was recorded, but the installments could not be saved or rolled back. Remove the disbursement transaction from the ledger before trying again." }); }
                }
                throw e;
            }
            return { added: schedule.installments.length as number };
        } catch (e) {
            console.error("Adding installments failed:", e);
            return fail(400, { error: reason(e) });
        }
    }
    ,updateInstallments: async ({ request, params }) => {
        const fields = await request.formData();
        const scheduleId = fields.get("scheduleId"), name = fields.get("name"), sourceAccountId = fields.get("sourceAccountId"), rows = fields.get("rows");
        if ([scheduleId,name,sourceAccountId,rows].some(v => typeof v !== "string")) return fail(400,{error:"Complete the installment plan details."});
        try {
            const installments = (rows as string).split(/\r?\n/).map((line) => line.trim()).filter(Boolean).map((line) => {
                const [dueDate, amount] = line.split(/[\t,]/).map((v) => v.trim());
                if (!/^\d{4}-\d{2}-\d{2}$/.test(dueDate ?? "") || !amount || !Number.isFinite(Number(amount))) throw new Error("Use one yyyy-mm-dd date and amount per line, separated by a comma.");
                return { dueDate, amount };
            });
            const firstInstallmentNumber=Number(fields.get("firstInstallmentNumber"));
            await updateInstallments(USER_ID,scheduleId as string,{name:name as string,sourceAccountId:sourceAccountId as string,accountId:params.id,firstInstallmentNumber,monthlyRatePercent:Number(fields.get("monthlyRatePercent")) || 0,installments});
            return { updated:true };
        } catch(e) { return fail(400,{error:reason(e,"Couldn't update this installment plan.")}); }
    },
    deleteInstallments: async ({ request }) => {
        const id=(await request.formData()).get("scheduleId");
        if(typeof id!=="string") return fail(400,{error:"Missing installment plan."});
        try { await deleteInstallments(USER_ID,id); return { deleted:true }; }
        catch(e) { return fail(400,{error:reason(e,"Couldn't delete this plan.")}); }
    },
    updateLoan: async ({ request, params }) => {
        const f=await request.formData();
        try {
            const loan = await getLoanDetails(USER_ID, params.id);
            await updateLoanDetails(USER_ID,{accountId:params.id,name:String(f.get("name")??""),bookedOn:String(f.get("bookedOn")??""),dailyRate:loan.dailyRate,principalOn:String(f.get("principalOn")??""),principal:Number(f.get("principal"))});
            return { loanUpdated:true };
        } catch(e) { return fail(400,{loanError:reason(e,"Couldn't update loan details.")}); }
    },
    deleteLoan: async ({ params }) => {
        try { await deleteLoan(USER_ID,params.id); throw redirect(303,"/accounts"); }
        catch(e) { if ((e as {status?:number}).status===303) throw e; return fail(400,{loanError:reason(e,"Couldn't delete this loan.")}); }
    }
};

/** The engine's reason for a rejected schedule (the API answers 400 with it), or a generic one. */
function reason(e: unknown, fallback = "Couldn't save the installments. Check that the API and engine are running."): string {
    try {
        const message = (e as Error).message;
        const body = JSON.parse(message.slice(message.indexOf("{")));
        if (body.status === 400 && body.message) return String(body.message).replace(/^invalid (?:schedule|transaction): /, "");
    } catch {
        // not a JSON error body
    }
    return fallback;
}
