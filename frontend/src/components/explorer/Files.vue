<script setup lang="ts"></script>

<template>
  <v-item-group selected-class="bg-primary">
    <v-container>
      <v-row>
        <v-col v-for="n in 10" :key="n" :cols="2">
          <v-item v-slot="{ isSelected, selectedClass, toggle }">
            <v-card
              max-width="10rem"
              :class="['pa-2 rounded-lg', selectedClass]"
              elevation="5"
              @click="toggle"
            >
              <!-- Upper part (file name and menu) -->
              <div
                class="tw-flex tw-flex-row tw-h-full tw-mb-3 tw-w-full tw-items-center tw-flex-wrap"
              >
                <p class="text-body-2 tw-grow">
                  {{ isSelected ? "Selected" : "Not Selected" }}
                </p>
                <v-menu>
                  <template v-slot:activator="{ props }">
                    <v-btn
                      density="compact"
                      icon="mdi-dots-vertical"
                      variant="plain"
                      v-bind="props"
                    ></v-btn>
                  </template>
                  <v-list>
                    <v-list-item @click="console.log('Download')">
                      <v-icon>mdi-download</v-icon> Download
                    </v-list-item>
                    <v-list-item @click="console.log('Details')">
                      <v-icon>mdi-information-outline</v-icon> Details
                    </v-list-item>
                    <v-list-item @click="console.log('Mark as favorite')">
                      <v-icon>mdi-star-outline</v-icon> Mark as favorite
                    </v-list-item>
                    <v-list-item @click="console.log('Delete')">
                      <v-icon>mdi-trash-can</v-icon> Delete
                    </v-list-item>
                  </v-list>
                </v-menu>
              </div>

              <div
                class="tw-flex tw-justify-center tw-items-center tw-mb-2 tw-w-full tw-h-16 tw-rounded-lg bg-grey-darken-3"
              >
                <v-icon icon="mdi-trash-can"></v-icon>
              </div>

              <!-- Bottom part (date) -->
              <p class="text-caption">1 day ago</p>
            </v-card>
          </v-item>
        </v-col>
      </v-row>
    </v-container>
  </v-item-group>
</template>

<style scoped></style>
