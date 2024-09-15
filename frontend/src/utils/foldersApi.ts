import axios from "axios";
import type Folder from "../models/folder";
import type FolderHierarchy from "../models/folderHierarchy";
import { useEventEmitterStore } from "../stores/eventEmitterStore";
import { FOLDER_UPDATED } from "../constants";
import { type FolderPatchRequest } from "../models/requestModel";

interface getFoldersResponse {
    folders: Folder[];
    hierarchies: FolderHierarchy[];
}

export async function getFolderList(folderCode: string): Promise<getFoldersResponse> {
    try {
        const response = await axios.get(`/api/folders/${folderCode}/folders`);
        return response.data as getFoldersResponse;
    } catch (error) {
        console.error(error);
    }

    return {
        folders: [],
        hierarchies: [],
    } as getFoldersResponse;
}

export async function getFavoriteFolders(): Promise<getFoldersResponse> {
    try {
        const response = await axios.get(`/api/folders/favorite`);
        return response.data as getFoldersResponse;
    } catch (error) {
        console.error(error);
    }

    return {
        folders: [],
        hierarchies: [],
    } as getFoldersResponse;
}

export async function createNewFolder(parentFolderCode: string, folderName: string): Promise<Folder> {
    try {
        const response = await axios.post(`/api/folders/${parentFolderCode}/folders`, {
            "folder_name": folderName,
        });
        const folder: Folder = response.data as Folder;
        const eventEmitter = useEventEmitterStore();
        eventEmitter.eventEmitter.emit(FOLDER_UPDATED);
        return folder;
    } catch (error) {
        console.error(error);
    }

    return {} as Folder;
}

export async function patchFolder(folderCode: string, patchRequest: FolderPatchRequest): Promise<Folder> {
    try {
        const response = await axios.patch(`/api/folders/${folderCode}`, patchRequest);
        const folder: Folder = response.data as Folder;
        // const eventEmitter = useEventEmitterStore();
        // eventEmitter.eventEmitter.emit(FOLDER_UPDATED);
        return folder;
    } catch (error) {
        console.error(error);
    }

    return {} as Folder;
}