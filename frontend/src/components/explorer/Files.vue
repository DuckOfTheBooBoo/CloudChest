<script setup lang="ts">
import { useRouter, useRoute } from "vue-router";
import { Ref, ref, onBeforeMount } from "vue";
import { MinIOFile } from "../../models/file";
import { getFilesFromPath } from "../../utils/filesApi";
import { formatDistance } from "date-fns";
import { fileDetailFormatter } from "../../utils/fileDetailFormatter";
import Filename from "../Filename.vue";
import File from "../File.vue";
import Folder from "../Folder.vue";

const fileDetailDialog = ref(false);
const fileList: Ref<MinIOFile[]> = ref([] as MinIOFile[]);
const folderList: Ref<string[]> = ref([] as string[]);

const path = ref("root");

const router = useRouter();
const route = useRoute();

onBeforeMount(async () => {
  const response = await getFilesFromPath(path.value);
  fileList.value = response.files;
  folderList.value = response.folders;
});
</script>

<template>
  <v-container>
    <v-row>
      <v-col v-for="file in [...fileList, ...folderList]" :key="file" :cols="2" ref="itemRefs">
        <File v-if="(typeof file !== 'string')" :file="file" />
        <Folder v-else :folderName="file" :path="path" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
