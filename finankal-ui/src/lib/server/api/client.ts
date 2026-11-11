const BASE_URL = "http://localhost:8080";

export async function api<T>(
    url: string,
    options?: RequestInit
): Promise<T> {

    const response = await fetch(`${BASE_URL}${url}`, {
        headers: {
            "Content-Type": "application/json",
        },
        ...options
    });

    if (!response.ok) {
        throw new Error(await response.text());
    }

    if (response.status === 204) return undefined as T;

    return response.json();
}
