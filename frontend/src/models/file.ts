import { parseISO } from "date-fns";
import Folder from "./folder";

export interface FileResponse {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  UserID: number;
  FolderID: number;
  FileName: string;
  FileCode: string;
  FileSize: number;
  FileType: string;  
  IsFavorite: boolean;
  Folder: Folder | null;
}

export class CloudChestFile {
  ID: number;
  CreatedAt: Date;
  UpdatedAt: Date;
  DeletedAt: Date | null;
  UserID: number;
  FolderID: number;
  FileName: string;
  FileCode: string;
  FileSize: number;
  FileType: string;
  IsFavorite: boolean;

  constructor(object: FileResponse) {
    this.ID = object.ID;
    this.CreatedAt = parseISO(object.CreatedAt);
    this.UpdatedAt = parseISO(object.UpdatedAt);
    this.DeletedAt = object.DeletedAt ? parseISO(object.DeletedAt) : null;
    this.UserID = object.UserID;
    this.FolderID = object.FolderID;
    this.FileName = object.FileName;
    this.FileCode = object.FileCode;
    this.FileSize = object.FileSize;
    this.FileType = object.FileType;
    this.IsFavorite = object.IsFavorite;
  }
}

export function parseFileResponse(response: FileResponse): CloudChestFile {
  return new CloudChestFile(response);
}
