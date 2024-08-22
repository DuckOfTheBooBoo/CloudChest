<script setup lang="ts">
import { type Ref, ref, inject } from "vue"
import Folder from "../models/folder";
import { getFolderList } from "../utils/foldersApi";

const props = defineProps<{
    folder: Folder;
    level: number;
}>();

const emits = defineEmits<{
    (e: 'click', folder: Folder): void;
}>();

const expandChild = ref<boolean>(false);
const isLoading = ref<boolean>(false);
const folderList: Ref<Folder[]> = ref([] as Folder[]);

const selectedFolder: Ref<Folder | null> | undefined = inject('selectedFolder');
const blacklistedFolder: Ref<Folder | undefined> | undefined = inject('blacklistedFolder');

async function fetchChildFolders(): Promise<void> {
    isLoading.value = true;
    try {
        const resp = await getFolderList(props.folder.Code);
        folderList.value = resp.folders;
        isLoading.value = false;
    } catch (error) {
        console.error(error)
    }
}

function toggleExpandChild(): void {
    if (expandChild.value) {
        expandChild.value = false;
        return
    }

    expandChild.value = true;
    fetchChildFolders();
}

function handleNested(folder: Folder): void {
    emits('click', folder);
}
</script>

<template>
    <div class="tw-flex tw-flex-col tw-my-1" v-bind="$attrs" v-if="blacklistedFolder && folder.Code !== blacklistedFolder?.Code">
        <div 
            :style="'margin-left: ' + level * 2 + 'rem'"
            class="tw-flex tw-justify-start tw-gap-3 tw-py-1 tw-rounded-lg hover:tw-bg-[#424242] tw-transition-colors tw-cursor-pointer"
            :class="{ 'tw-bg-[#424242]': selectedFolder?.Code === folder.Code }"
            @click="emits('click', folder)">
            <v-btn :icon="expandChild ? 'mdi-chevron-down' : 'mdi-chevron-right'" :class="{ 'tw-invisible': !folder.HasChild }" @click="toggleExpandChild" variant="text" density="compact"></v-btn>
            <v-icon>mdi-folder</v-icon>
            <span>{{ folder.Name }}</span>
        </div>
        <v-expand-transition>
            <div v-show="expandChild">
                <v-progress-circular v-if="isLoading" indeterminate></v-progress-circular>
                <div v-else v-for="folder in folderList" :key="folder.Code">
                    <FolderListView @click="handleNested" :folder="folder" :level="level + 1" />
                </div>
            </div>
        </v-expand-transition>
    </div>
</template>