import { parseISO } from "date-fns";

export interface FileResponse {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  UserID: number;
  FileName: string;
  FileSize: number;
  FileType: string;
  StoragePath: string;
}

export class MinIOFile {
  ID: number;
  CreatedAt: Date;
  UpdatedAt: Date;
  DeletedAt: Date | null;
  UserID: number;
  FileName: string;
  FileSize: number;
  FileType: string;
  StoragePath: string;
  constructor(object: FileResponse) {
    this.ID = object.ID;
    this.CreatedAt = parseISO(object.CreatedAt);
    this.UpdatedAt = parseISO(object.UpdatedAt);
    this.DeletedAt = object.DeletedAt ? parseISO(object.DeletedAt) : null;
    this.UserID = object.UserID;
    this.FileName = object.FileName;
    this.FileSize = object.FileSize;
    this.FileType = object.FileType;
    this.StoragePath = object.StoragePath;
  }
}

export function parseFileResponse(response: FileResponse): MinIOFile {
  return new MinIOFile(response);
}
