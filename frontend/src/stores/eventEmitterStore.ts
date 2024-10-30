import { defineStore } from "pinia";
import EventEmitter from "eventemitter3";
import { CloudChestFile } from "../models/file";
import type Folder from "../models/folder";

interface Events {
  FILE_UPDATED: (file: CloudChestFile) => void;
  FILE_DELETED_TEMP: (file: CloudChestFile) => void;
  FILE_DELETED_PERM: (file: CloudChestFile) => void;
  FILE_ADDED: (file: CloudChestFile) => void;
  FOLDER_UPDATED: (folder: Folder) => void;
  FOLDER_DELETED_TEMP: (folder: Folder) => void;
  FOLDER_DELETED_PERM: (deletedObjects: {deleted_files: string[], deleted_folders: string[]}) => void;
  FOLDER_ADDED: (folder: Folder) => void;
}

class MyEmitter extends EventEmitter<Events> {}

export const useEventEmitterStore = defineStore("eventEmitter", {
  state: () => ({
    eventEmitter: new MyEmitter(),
  }),
  getters: {
    getEventEmitter: (state) => state.eventEmitter,
  },
});
