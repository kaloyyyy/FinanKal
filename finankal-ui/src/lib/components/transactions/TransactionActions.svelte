<script lang="ts">
    interface Props {
        canSubmit: boolean;
        loading?: boolean;
        onCancel?: () => void;
    }

    let {
        canSubmit,
        loading = false,
        onCancel
    }: Props = $props();
</script>

<div class="flex items-center justify-end gap-3 border-t border-slate-800 pt-6">

    <button
            type="button"
            onclick={() => onCancel?.()}
            class="rounded-xl border border-slate-700 bg-slate-900 px-6 py-3 font-medium text-slate-300 transition hover:border-slate-600 hover:bg-slate-800 hover:text-white"
    >
        Cancel
    </button>

    <button
            type="submit"
            disabled={!canSubmit || loading}
            class="rounded-xl bg-blue-600 px-6 py-3 font-semibold text-white transition hover:bg-blue-500 disabled:cursor-not-allowed disabled:bg-slate-700 disabled:text-slate-400"
    >
        {#if loading}
            Creating Transaction...
        {:else}
            Create Transaction
        {/if}
    </button>

</div>

