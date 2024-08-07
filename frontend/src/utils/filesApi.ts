import axios from "axios";
import { CloudChestFile, type FileResponse } from "../models/file";
import { type PresignedURL } from "../models/presignedUrl";
import { useEventEmitterStore } from "../stores/eventEmitterStore";
import { FILE_UPDATED } from "../constants";

export async function getFilesFromCode(folderCode: string): Promise<CloudChestFile[]> {
  try {
    const response = await axios.get("http://localhost:3000/api/files/" + folderCode, {
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
    const response = await axios.get("http://localhost:3000/api/files", {
      params: {
        trashCan: true,
        path: 'a'
      }
    });
    return { files: response.data.files as CloudChestFile[] };
  } catch (error) {
    console.error(error);
  }
  return { files: [] };
}

export async function getFavoriteFiles(): Promise<{files: CloudChestFile[]}> {
  try {
    const response = await axios.get("http://localhost:3000/api/files", {
      params: {
        favorite: true,
        path: 'a'
      }
    });
    return { files: response.data.files as CloudChestFile[] };
  } catch (error) {
    console.error(error);
  }
  return { files: [] };
}

export async function trashFile(file: CloudChestFile, isTrashFile: boolean): Promise<void> {
  const url: string = `http://localhost:3000/api/files/${file.ID}?trash=${isTrashFile}`;
  try {
    await axios.delete(url);
    const eventEmitter = useEventEmitterStore();
    eventEmitter.eventEmitter.emit(FILE_UPDATED);
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
    await axios.put(`http://localhost:3000/api/files/${file.ID}`, body);
    const eventEmitter = useEventEmitterStore();
    eventEmitter.eventEmitter.emit(FILE_UPDATED);
    return true;
  } catch (error) {
    console.error(error);
  }

  return false;
}

export async function downloadFile(fileID: number): Promise<PresignedURL> {
  try {
    const response = await axios.get(`http://localhost:3000/api/files/download/${fileID}`);
    return response.data as PresignedURL;
  } catch (error) {
    console.error(error);
  }

  return {} as PresignedURL;
}