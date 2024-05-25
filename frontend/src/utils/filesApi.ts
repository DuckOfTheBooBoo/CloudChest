import axios from "axios";
import { MinIOFile, FileResponse } from "../models/file";

interface FilesResponse {
  files: FileResponse[];
}

export async function getAllFiles(): Promise<MinIOFile[]> {
  try {
    const response = await axios.get("http://localhost:3000/api/files");
    if (response.data.hasOwnProperty("files")) {
      const filesResponse: FilesResponse = response.data as FilesResponse;
      const files: MinIOFile[] = filesResponse.files.map(
        (fileResponse) => new MinIOFile(fileResponse)
      );
      return files;
    }
  } catch (error: any) {
    console.error(error);
  }
  return [] as MinIOFile[];
}
