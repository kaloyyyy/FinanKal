// stores/account.ts

import { writable } from "svelte/store";
import type { AccountDto } from "$lib/types/account";

export const account = writable<AccountDto | null>(null);

export const loading = writable(false);