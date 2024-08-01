<script setup lang="ts">
import { ref, computed } from "vue";
import { formatDistance } from "date-fns";
import { fileDetailFormatter } from "../utils/fileDetailFormatter";
import Folder from "../models/folder";
import Filename from "./Filename.vue";

const props = defineProps<{
  folder: Folder;
  parentPath: string;
  makeRequest: Function;
}>();

const decodedParentPath = computed(() => decodeURIComponent(props.parentPath))
const fullPath = computed(() => decodedParentPath.value === '/' ? `/${props.folder.DirName}` : `${decodedParentPath.value}/${props.folder.DirName}`);
const folderName = computed(() => props.folder.DirName);

const folderDetailDialog = ref(false);

function doRequest() {
  props.makeRequest(fullPath.value);
}
</script>

<template>
  <v-tooltip :text="folderName" location="bottom">
    <template v-slot:activator="{ props: tltpProps }">
      <v-card max-width="10rem" class="pa-2 rounded-lg" hover @click="doRequest" v-bind="tltpProps">
        <!-- Upper part (file name and menu) -->
        <div class="tw-flex tw-flex-row tw-h-full tw-mb-3 tw-w-full tw-items-center tw-justify-between">
          <Filename :filename="folderName" />
          <v-menu>
            <template v-slot:activator="{ props: menuProps }">
              <v-btn density="compact" icon="mdi-dots-vertical" variant="plain" v-bind="menuProps"></v-btn>
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

        <div class="tw-flex tw-justify-center tw-items-center tw-mb-2 tw-w-full tw-h-16 tw-rounded-lg bg-grey-darken-3">
          <v-icon icon="mdi-trash-can"></v-icon>
        </div>

        <!-- Bottom part (date) -->
        <p class="text-caption">
          {{ formatDistance(folder.CreatedAt, new Date(), { addSuffix: true }) }}
        </p>
      </v-card>
    </template>
  </v-tooltip>
</template>

  <script setup></script>
