<script lang="ts">
    import { Landmark, Wallet, CreditCard, Receipt, HandCoins, Folder, ArrowDownLeft, ArrowUpRight, Banknote } from "lucide-svelte";
    import { bankBadge } from "$lib/utils/bankColors";

    let { name, type }: { name: string; type: string } = $props();

    // Only accounts that hold or owe money carry a bank's colours; categories like "Atome Purchases" don't.
    const MONEY = ["ASSET", "CREDIT_CARD", "LIABILITY", "RECEIVABLE", "PAYABLE"];
    const badge = $derived(MONEY.includes(type) ? bankBadge(name) : null);

    // The account's kind: a corner glyph on a bank badge, or the whole icon otherwise.
    const Icon = $derived(
        type === "CREDIT_CARD" ? CreditCard
        : type === "LIABILITY" || type === "PAYABLE" ? Receipt
        : type === "RECEIVABLE" ? HandCoins
        : type === "ASSET" ? (/^cash\b/i.test(name) ? Banknote : badge?.wallet ? Wallet : Landmark)
        : type === "INCOME" ? ArrowDownLeft
        : type === "EXPENSE" ? ArrowUpRight
        : Folder
    );

    // Categories are coloured by direction: money coming in, money going out.
    const tint =
        $derived(type === "INCOME" ? "bg-emerald-500/15 text-emerald-300"
        : type === "EXPENSE" ? "bg-rose-500/15 text-rose-300"
        : "bg-slate-800 text-slate-300");
</script>

{#if badge}
    <!-- The bank's colour and letters, matching its swatch in the position cards -->
    <span class="relative inline-flex h-10 w-10 shrink-0" aria-hidden="true">
        <span
                class="flex h-10 w-10 items-center justify-center rounded-xl text-[13px] font-bold tracking-tight [text-shadow:0_1px_2px_rgba(0,0,0,0.35)]"
                style:background={badge.background}
                style:color={badge.ink}
        >
            {badge.mark}
        </span>
        <span class="absolute -right-1 -bottom-1 flex h-[18px] w-[18px] items-center justify-center rounded-full bg-slate-950 text-slate-300 ring-2 ring-slate-950">
            <Icon class="h-2.5 w-2.5" strokeWidth={2.5} />
        </span>
    </span>
{:else}
    <span class={`flex h-10 w-10 shrink-0 items-center justify-center rounded-xl ${tint}`} aria-hidden="true">
        <Icon class="h-4 w-4" />
    </span>
{/if}
