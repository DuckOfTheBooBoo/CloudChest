import { defineStore } from "pinia";
import EventEmitter from "eventemitter3";

export const useEventEmitterStore = defineStore("eventEmitter", {
  state: () => ({
    eventEmitter: new EventEmitter(),
  }),
  getters: {
    getEventEmitter: (state) => state.eventEmitter,
  },
});
