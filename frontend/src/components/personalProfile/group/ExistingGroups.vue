<script setup>
import { LIMITS } from '@/helpers/limits';
import { ref, computed } from 'vue';

import { deletePostGroup } from '@/api/common/friends.js';
import { addNotification } from '@/data/notifications';
import { fullName } from './groupHelpers.js';
import GroupCard from './GroupCard.vue';
import EditGroupModal from './EditGroupModal.vue';

const props = defineProps({
    groups: {
        type: Array,
        default: () => []
    }
});

const emit = defineEmits(['changed']);

const groupSearchQuery = ref('');

const allCards = computed(() =>
    props.groups.map((group) => {
        const users = group.Users || [];

        return {
            raw: group,
            id: group.ID,
            name: group.name,
            firstName: fullName(users[0]),
            lastName: users.length > 1 ? fullName(users[users.length - 1]) : '',
            extraCount: users.length > 2 ? users.length - 2 : 0
        };
    })
);

const cards = computed(() => {
    const query = groupSearchQuery.value.trim().toLowerCase();

    if (!query) {
        return allCards.value;
    }

    return allCards.value.filter((card) =>
        (card.name || '').toLowerCase().includes(query)
    );
});

const editingGroup = ref(null);

function openEdit(card) {
    editingGroup.value = card.raw;
}

function closeEdit() {
    editingGroup.value = null;
}

function onSaved() {
    closeEdit();
    emit('changed');
}

async function onDelete(id) {
    try {
        const result = await deletePostGroup(id);

        if (!result.status) {
            addNotification(result.message, 'error');
            return;
        }

        addNotification('group deleted!', 'success');
        emit('changed');
    } catch (err) {
        addNotification(err, 'error');
    }
}

</script>

<template>
    <div class="existing-groups">
        <input
            v-model="groupSearchQuery"
            :maxlength="LIMITS.search"
            type="text"
            placeholder="Search groups by name..."
            class="group-search-input"
        />

        <div v-if="!cards.length" class="state-message">
            {{ groupSearchQuery.trim() ? 'No groups match your search.' : 'No groups yet.' }}
        </div>

        <div v-else class="groups-grid">
            <GroupCard
                v-for="card in cards"
                :key="card.id"
                :card="card"
                @edit="openEdit(card)"
                @delete="onDelete(card.id)"
            />
        </div>

        <EditGroupModal
            v-if="editingGroup"
            :group="editingGroup"
            @close="closeEdit"
            @saved="onSaved"
        />
    </div>
</template>

<style scoped>
* {
    box-sizing: border-box;
}

.existing-groups {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.group-search-input {
    padding: 12px 14px;
    border: 2px solid #1a1a1a;
    border-radius: 8px;
    font-size: 15px;
    max-width: 320px;
}

.groups-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 14px;
}

.state-message {
    font-family: "JetBrains Mono", monospace;
    font-size: 12px;
    color: #55554f;
}
</style>