<script lang="ts">
    import type {LedgerEntryDto} from "$lib/types/ledger";

    interface Props {
        entries: LedgerEntryDto[];
    }

    let {entries}: Props = $props();
</script>

<table class="table">

    <thead>

    <tr>

        <th>Date</th>

        <th>Description</th>

        <th>Type</th>

        <th>Amount</th>

    </tr>

    </thead>

    <tbody>

    {#each entries as entry}

        <tr>

            <td>{entry.createdAt}</td>

            <td>{entry.description}</td>

            <td>{entry.type}</td>

            <td>{entry.amount}</td>

        </tr>

    {/each}

    </tbody>

</table>