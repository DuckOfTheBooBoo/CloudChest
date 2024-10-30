<script setup lang="ts">
import { computed, type ComputedRef } from 'vue';
import { CloudChestFile } from '../models/file';

const props = defineProps<{
    file?: CloudChestFile | null;
    visible: boolean;
}>()

const emit = defineEmits<{
  (e: "on:close"): void;
}>();

const isPreviewable: ComputedRef<boolean | undefined> = computed(() => props.file?.FileType.includes('image/') || props.file?.FileType.includes('video/') && props.file?.IsPreviewable)
const fileURL: ComputedRef<string | undefined> = computed(() => `/api/files/${props.file?.FileCode}/download`)
</script>

<template>
    <v-overlay v-bind:model-value="visible" scroll-strategy="block">
      <v-toolbar class="tw-w-screen" density="comfortable">
        <v-toolbar-title>{{ file?.FileName }}</v-toolbar-title>

        <v-spacer></v-spacer>

        <v-btn @click="emit('on:close')" icon>
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar>

      <div class="tw-py-6 tw-flex tw-justify-center tw-items-center tw-drop-shadow-xl">
        <v-img v-if="isPreviewable && file?.FileType.includes('image/')" :src="fileURL"
          class="tw-h-[calc(100dvh-100px)]">
          <template v-slot:placeholder>
            <div class="d-flex align-center justify-center fill-height">
              <v-progress-circular color="grey-lighten-4" indeterminate></v-progress-circular>
            </div>
          </template>
        </v-img>
        <media-controller class="tw-h-[calc(100dvh-100px)]" v-else-if="isPreviewable && file?.FileType.includes('video/')">
          <hls-video :src="`/api/hls/${file?.FileCode}/masterPlaylist`" slot="media"
            crossorigin muted></hls-video>
          <media-loading-indicator slot="centered-chrome" noautohide></media-loading-indicator>
          <media-control-bar>
            <media-play-button></media-play-button>
            <media-seek-backward-button></media-seek-backward-button>
            <media-seek-forward-button></media-seek-forward-button>
            <media-mute-button></media-mute-button>
            <media-volume-range></media-volume-range>
            <media-time-range></media-time-range>
            <media-time-display showduration remaining></media-time-display>
            <media-playback-rate-button></media-playback-rate-button>
            <media-fullscreen-button></media-fullscreen-button>
          </media-control-bar>
        </media-controller>
        <p v-else class="tw-text-2xl">This file is does not have a preview.</p>
      </div>
    </v-overlay>
</template>