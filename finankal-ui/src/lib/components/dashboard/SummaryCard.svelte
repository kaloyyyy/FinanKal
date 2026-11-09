<script lang="ts">
    import type { Wallet } from "lucide-svelte";

    type Props = {
        title: string;
        value: string;
        description?: string;
        icon?: typeof Wallet;
        variant?: "default" | "positive" | "negative";
    };

    let {
        title,
        value,
        description,
        icon: Icon,
        variant = "default"
    }: Props = $props();

    const valueColor = $derived({
        default: "text-white",
        positive: "text-emerald-400",
        negative: "text-red-400"
    }[variant]);
</script>

<div
        class="rounded-2xl border border-slate-800 bg-slate-900/80 p-6 shadow-xl backdrop-blur transition duration-200 hover:border-slate-700 hover:bg-slate-900"
>
    <div class="flex items-start justify-between gap-4">
        <div class="min-w-0">
            <p class="text-sm font-medium text-slate-400">
                {title}
            </p>

            <h2 class="mt-3 text-3xl font-bold tracking-tight tabular-nums {valueColor}">
                {value}
            </h2>

            {#if description}
                <p class="mt-3 text-xs text-slate-500">
                    {description}
                </p>
            {/if}
        </div>

        {#if Icon}
            <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-slate-800 shadow-inner {valueColor}">
                <Icon class="h-6 w-6" />
            </div>
        {/if}
    </div>
</div>
