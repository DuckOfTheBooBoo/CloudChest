import axios from "axios";
import Folder from "../models/folder";
import { useEventEmitterStore } from "../stores/eventEmitterStore";
import { FILE_UPDATED, FOLDER_UPDATED } from "../constants";

export async function getFolderList(folderCode: string): Promise<Folder[]> {
    try {
        const response = await axios.get("http://localhost:3000/api/folders/" + folderCode);
        return response.data as Folder[];
    } catch (error) {
        console.error(error);
    }

    return [];
}

export async function createNewFolder(parentFolderCode: string, folderName: string): Promise<Folder> {
    try {
        const response = await axios.post("http://localhost:3000/api/folders/" + parentFolderCode, {
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