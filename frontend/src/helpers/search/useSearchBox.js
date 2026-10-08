import { ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export const SEARCH_PATH = '/search';

export function useSearchBox() {
    const route = useRoute();
    const router = useRouter();

    function currentQuery() {
        if (route.path === SEARCH_PATH && typeof route.query.q === 'string') {
            return route.query.q;
        }

        return '';
    }

    const searchQuery = ref(currentQuery());

    watch(
        () => [route.path, route.query.q],
        () => {
            searchQuery.value = currentQuery();
        }
    );

    function submitSearch() {
        const query = searchQuery.value.trim();

        if (!query) {
            return;
        }

        router.push({
            path: SEARCH_PATH,
            query: { q: query }
        });
    }

    function clearSearch() {
        searchQuery.value = '';
    }

    return {
        searchQuery,
        submitSearch,
        clearSearch
    };
}
