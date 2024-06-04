<script setup lang="ts">
import { useRoute, useRouter } from "vue-router";
import { watch, Ref, ref, onBeforeMount, onBeforeUnmount, onMounted } from "vue";
import { MinIOFile } from "../../models/file";
import { getFilesFromPath } from "../../utils/filesApi";
import File from "../File.vue";
import Folder from "../Folder.vue";
import FolderModel from "../../models/folder";

const fileList: Ref<MinIOFile[]> = ref([] as MinIOFile[]);
const folderList: Ref<FolderModel[]> = ref([] as FolderModel[]);

const route = useRoute();
const router = useRouter();
const path = ref('/');

window.addEventListener('popstate', () => {
  console.log(path.value, route.path)
  if(path.value) {
    path.value = route.path; // Update path based on current route
  } else {
    path.value = '/';
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('popstate', () => {});
});

onBeforeMount(async () => {
  path.value = decodeURIComponent(route.query.path as string);
  const response = await getFilesFromPath(path.value);
  fileList.value = response.files;
  folderList.value = response.folders;
});

onMounted(() => {
  console.log(route.fullPath)
})

async function makeRequest(pathParam: string): Promise<void> {
  path.value = pathParam
  const response = await getFilesFromPath(path.value);
  fileList.value = response.files;
  folderList.value = response.folders;
  router.push({ path: '/explorer/files', query: { path: encodeURIComponent(path.value) }})
}
</script>

<template>
  <v-container>
    <v-row>
      <v-col v-for="folder in folderList" :key="folder" :cols="2">
        <Folder :folder="folder" :parent-path="decodeURIComponent(path)" :make-request="makeRequest"/>
      </v-col>
      <v-col v-for="file in fileList" :key="file" :cols="2">
        <File :file="file" />
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped></style>
