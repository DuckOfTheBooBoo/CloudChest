import axios from "axios";
import Folder from "../models/folder";
import { useEventEmitterStore } from "../stores/eventEmitterStore";
import { FILE_UPDATED } from "../constants";

export async function getFolderList(folderCode: string): Promise<Folder[]> {
    try {
        const response = await axios.get("http://localhost:3000/api/folders/" + folderCode);
        return response.data as Folder[];
    } catch (error) {
        console.error(error);
    }

    return [];
}