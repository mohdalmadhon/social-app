import { safeSearch } from '@/helpers/limits';

export async function getSuggestedLocations(searchValue = "") {
    const resp = await fetch(
        `/api/location/search?q=${encodeURIComponent(safeSearch(searchValue))}`,
        {
            method: "GET"
        }
    );

    if (!resp.ok) {
        throw new Error("could not fetch locations");
    }

    const data = await resp.json();

    return data;
}