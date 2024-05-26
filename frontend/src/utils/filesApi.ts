import axios from "axios";
import { MinIOFile, FileResponse } from "../models/file";

interface FilesResponse {
  files: FileResponse[] | null;
  folders: string[] | null;
}

interface ReturnValue {
  files: MinIOFile[];
  folders: string[];
}

export async function getFilesFromPath(path: string): Promise<ReturnValue> {
  try {
    const response = await axios.get("http://localhost:3000/api/files/"+path);
    if (response.data.hasOwnProperty("files")) {
      const filesResponse: FilesResponse = response.data as FilesResponse;
      const files: MinIOFile[] = filesResponse.files?.map(
        (fileResponse) => new MinIOFile(fileResponse)
      )!;
      const folders = filesResponse.folders
      return {
        files,
        folders
      } as ReturnValue;
    }
  } catch (error: any) {
    console.error(error);
  }
  return { files: [], folders: [] } as ReturnValue;
}
