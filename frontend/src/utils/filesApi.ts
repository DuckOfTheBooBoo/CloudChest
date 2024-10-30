import axios from "axios";
import { CloudChestFile, type FileResponse } from "../models/file";
import { type PresignedURL } from "../models/presignedUrl";
import { useEventEmitterStore } from "../stores/eventEmitterStore";
import { type FilePatchRequest } from "../models/requestModel";

export async function getFilesFromCode(folderCode: string): Promise<CloudChestFile[]> {
  try {
    const response = await axios.get(`/api/folders/${folderCode}/files`, {
      params: {
        trashCan: false
      }
    });
    const filesResponse: FileResponse[] = response.data as FileResponse[];
    const files: CloudChestFile[] = filesResponse.map(
      (fileResponse) => new CloudChestFile(fileResponse)
    )!;
    return files;
  } catch (error: any) {
    console.error(error);
  }
  return [];
}

export async function getTrashCan(): Promise<{files: CloudChestFile[]}> {
  try {
    const response = await axios.get("/api/files/trashcan");
    return { files: response.data as CloudChestFile[] };
  } catch (error) {
    console.error(error);
  }
  return { files: [] };
}

export async function getFavoriteFiles(): Promise<{files: CloudChestFile[]}> {
  try {
    const response = await axios.get("/api/files/favorite");
    return { files: response.data as CloudChestFile[] };
  } catch (error) {
    console.error(error);
  }
  return { files: [] };
}

export async function trashFile(file: CloudChestFile, isTrashFile: boolean): Promise<void> {
  const evStore = useEventEmitterStore();
  const url: string = `/api/files/${file.ID}?trash=${isTrashFile}`;
  try {
    await axios.delete(url);
    if (isTrashFile) {
      evStore.getEventEmitter.emit("FILE_DELETED_TEMP", file);
    } else {
      evStore.getEventEmitter.emit("FILE_DELETED_PERM", file);
    }
  } catch (error) {
    console.error(error);
  }
}

export async function emptyTrashCan(): Promise<void> {
  const url: string = `/api/files`;
  try {
    await axios.delete(url);
  } catch (error) {
    console.error(error);
  }
}

export async function updateFile(file: CloudChestFile, isRestoreFile: boolean): Promise<boolean> {
  const body: {
    file_name: string,
    is_favorite: boolean,
    is_restore: boolean } = {
    file_name: file.FileName,
    is_favorite: file.IsFavorite,
    is_restore: isRestoreFile,
  };

  try {
    await axios.put(`/api/files/${file.ID}`, body);
    return true;
  } catch (error) {
    console.error(error);
  }

  return false;
}

export async function patchFile(file: CloudChestFile, patchRequest: FilePatchRequest): Promise<void> {
  const evStore = useEventEmitterStore()
  try {
    const response = await axios.patch(`/api/files/${file.ID}`, patchRequest);
    evStore.getEventEmitter.emit("FILE_UPDATED", response.data as CloudChestFile);
  } catch (error) {
    console.error(error);
  }
}

export async function downloadFile(fileCode: string): Promise<PresignedURL> {
  try {
    const response = await axios.get(`/api/files/${fileCode}/download`);
    return response.data as PresignedURL;
  } catch (error) {
    console.error(error);
  }

  return {} as PresignedURL;
}
