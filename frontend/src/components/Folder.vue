<script setup lang="ts">
import { ref } from "vue";
import Folder from "../models/folder";
import Filename from "./Filename.vue";

defineProps<{
  folder: Folder;
  isSelected: boolean
}>();

const emit = defineEmits<{
  (e: "folderCode:change", newFolderCode: string): void
}>();

</script>

<template>
  <v-tooltip :text="folder.Name" location="bottom">
    <template v-slot:activator="{ props: tltpProps }">
      <div
        class="tw-flex tw-flex-col tw-items-center tw-max-w-15rem hover:tw-cursor-pointer hover:tw-bg-[#424242] tw-rounded-xl tw-pb-2 tw-transition-[background-color]"
        @dblclick="emit('folderCode:change', folder.Code)"
        :class="{ 'tw-bg-[#424242]': isSelected }" v-bind="tltpProps">
        <v-icon class="tw-text-9xl !important">mdi-folder</v-icon>
        <Filename :filename="folder.Name" />
      </div>

    </template>
  </v-tooltip>
</template>

<script setup></script>
