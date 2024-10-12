import { defineStore } from "pinia";
import axios, { type CancelTokenSource, type AxiosRequestConfig, type AxiosProgressEvent } from "axios";
import type RequestData from "../models/requestData";
import { reactive } from "vue";
import { useEventEmitterStore } from "./eventEmitterStore"
import { FILE_UPDATED } from "../constants";
import { CloudChestFile, FileResponse } from "../models/file";

export const useAxiosManagerStore = defineStore("axiosManager", {
  state: () => {
    return {
      ongoingRequests: reactive<RequestData[]>([]),
      eventEmitter: useEventEmitterStore(),
    }
  },
  actions: {
    generateId(): string {
      return Math.random().toString(36).slice(2, 9);
    },
    addUploadRequest(file: File, folderCode: string, config?: AxiosRequestConfig): RequestData {
      const cancelToken: CancelTokenSource = axios.CancelToken.source();
      const id = this.generateId();
      const requestConfig: AxiosRequestConfig = {
        ...config,
        cancelToken: cancelToken.token,
        onUploadProgress: (progressEvent: AxiosProgressEvent) => {
          const requestData: RequestData = this.ongoingRequests.find((request) => request.id === id) as RequestData;
          if (requestData && progressEvent.total !== undefined) {
            requestData.progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          }
        }
      }

      const request = axios.postForm(`/api/folders/${folderCode}/files`, {
        file: file,
      }, requestConfig);

      const requestData: RequestData = {
        id, filename: file.name, request, cancelToken, progress: 0
      };

      this.ongoingRequests.push(requestData);

      request
        .then((resp) => {
          console.log(resp)
          if (resp.status === 201) {
            this.eventEmitter.getEventEmitter.emit("FILE_ADDED", new CloudChestFile(resp.data as FileResponse))
          }
        })
        .catch(err => console.error(err))
      return requestData;
    },

    cancelRequest(id: string): void {
      const requestData = this.ongoingRequests.find(req => req.id === id);
      if (requestData) {
        requestData.cancelToken.cancel(`Request with ID ${id} canceled.`);
        this.removeRequest(id);
      }
    },
    removeRequest(id: string): void {
      const index = this.ongoingRequests.findIndex(req => req.id === id);
      if (index !== -1) {
        this.ongoingRequests.splice(index, 1);
      }
    },
  }
});
