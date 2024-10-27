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

export async function getDeletedFolders(): Promise<getFoldersResponse> {
    try {
        const response = await axios.get(`/api/folders/trashcan`);
        return response.data as getFoldersResponse;
    } catch (error) {
        console.error(error);
    }

    return {} as getFoldersResponse;
}

export async function createNewFolder(parentFolderCode: string, folderName: string): Promise<void> {
    const evStore = useEventEmitterStore();
    try {
        const response = await axios.post(`/api/folders/${parentFolderCode}/folders`, {
            "folder_name": folderName,
        });
        const folder: Folder = response.data as Folder;
        evStore.getEventEmitter.emit("FOLDER_ADDED", folder)
    } catch (error) {
        console.error(error);
    }
}

export async function patchFolder(folderCode: string, patchRequest: FolderPatchRequest): Promise<void> {
    const evStore = useEventEmitterStore();
    try {
        const response = await axios.patch(`/api/folders/${folderCode}`, patchRequest);
        const folder: Folder = response.data["Folder"] as Folder;
        const code: string = response.data["Code"] as string;
        folder.Code = code;
        evStore.getEventEmitter.emit("FOLDER_UPDATED", folder)
    } catch (error) {
        console.error(error);
    }
}

export async function deleteFolderTemp(folderCode: string): Promise<void> {
    try {
        await axios.delete(`/api/folders/${folderCode}?trash=true`);
    } catch (error) {
        console.error(error);
    }
}

export async function deleteFolderPermanent(folderCode: string): Promise<{deleted_files: string[], deleted_folders: string[]}> {
    try {
        const response = await axios.delete(`/api/folders/${folderCode}?trash=false`);
        return {
            deleted_files: response.data.deleted_files,
            deleted_folders: response.data.deleted_folders
        }
    } catch (error) {
        console.error(error);
    }

    return {} as {deleted_files: string[], deleted_folders: string[]};
}