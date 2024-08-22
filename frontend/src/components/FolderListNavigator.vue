<script setup lang="ts">
import type Folder from "../models/folder";
import { getFolderList } from "../utils/foldersApi";
import FolderListView from "./FolderListView.vue";
import { onMounted, type Ref, ref, provide } from "vue";

const props = defineProps<{blacklistedFolder?: Folder}>();

const folders: Ref<Folder[]> = ref([] as Folder[]);
const selectedFolder: Ref<Folder | null> = ref(null);

provide('selectedFolder', selectedFolder);
if (props.blacklistedFolder !== undefined) {
    provide('blacklistedFolder', props.blacklistedFolder);
}

const emits = defineEmits<{
    (e: "nav:cancel"): void;
    (e: "nav:move", folder: Folder | null): void
}>();

onMounted(async () => {
    const resp = await getFolderList("root");
    folders.value = resp.folders;
})

function handleFolderSelect(folder: Folder): void {
    selectedFolder.value = folder;
}
</script>

<template>
    <v-card>
        <v-card-title primary-title>
            Move to <span class="">{{ selectedFolder?.Name }}</span>
        </v-card-title>

        <v-card-text>
            <div>
                <FolderListView @click="handleFolderSelect" v-for="folder in folders" :folder="folder" :key="folder.Code" :level="0" />
            </div>
        </v-card-text>

        <v-card-actions>
            <v-btn color="error" variant="outlined" @click="emits('nav:cancel')">Cancel</v-btn>
            <v-btn color="primary" :disabled="!selectedFolder" @click="emits('nav:move', selectedFolder)">Move</v-btn>
        </v-card-actions>
    </v-card>
</template>