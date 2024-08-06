<script setup lang="ts">
import { ref } from "vue";
import { formatDistance } from "date-fns";
import Folder from "../models/folder";
import Filename from "./Filename.vue";

defineProps<{
  folder: Folder;
  isSelected: boolean
}>();

const emit = defineEmits<{
  (e: "folderCode:change", newFolderCode: string): void
}>();

const folderDetailDialog = ref(false);
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

        <v-menu>
          <template v-slot:activator="{ props: menuProps }">
            <!-- <v-btn density="compact" icon="mdi-dots-vertical" variant="plain" v-bind="menuProps"></v-btn> -->
          </template>
          <v-list>
            <v-list-item @click="console.log('Download')">
              <v-icon>mdi-download</v-icon> Download
            </v-list-item>
            <!-- <v-list-item @click="() => {}">
              DETAILS DIALOG
              <v-dialog
                activator="parent"
                max-width="30rem"
                v-model="folderDetailDialog"
              >
                <template v-slot:default="{ isActive: _ }">
                  <v-card>
                    <v-card-title>
                      <div
                        class="tw-flex tw-flex-row tw-justify-between tw-items-center tw-px-2"
                      >
                        <p>File details</p>
                        <v-btn
                          icon="mdi-close"
                          variant="flat"
                          @click="folderDetailDialog = false"
                        ></v-btn>
                      </div>
                    </v-card-title>
                    <v-card-text>
                      <div
                        v-for="(value, index) in Object.entries(
                          fileDetailFormatter(file)
                        )"
                        :key="index"
                        class="tw-flex tw-flex-col"
                      >
                        <div class="tw-flex tw-flex-row tw-justify-start">
                          <p class="tw-w-1/2">{{ value[0] }}</p>
                          <p class="tw-w-1/2">{{ value[1] }}</p>
                        </div>
                        <v-divider
                          class="tw-my-2"
                          v-if="
                            index !=
                            Object.entries(fileDetailFormatter(file)).length - 1
                          "
                        ></v-divider>
                      </div>
                    </v-card-text>
                  </v-card>
                </template>
              </v-dialog>
  
              <v-icon>mdi-information-outline</v-icon> Details
            </v-list-item> -->
            <v-list-item @click="console.log('Mark as favorite')">
              <v-icon>mdi-star-outline</v-icon> Mark as favorite
            </v-list-item>
            <v-list-item @click="console.log('Delete')">
              <v-icon>mdi-trash-can</v-icon> Delete
            </v-list-item>
          </v-list>
        </v-menu>
      </div>

    </template>
  </v-tooltip>
</template>

<script setup></script>
