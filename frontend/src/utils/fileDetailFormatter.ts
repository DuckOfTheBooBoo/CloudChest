import { MinIOFile } from "../models/file";
import { format } from "date-fns";



export default function fileDetailFormatter(file: MinIOFile): Object {
    return {
        "ID": file.ID,
        "File name": file.FileName,
        "File type": file.FileType,
        "File size": file.FileSize,
        "Location": file.StoragePath,
        "Created at": format(file.CreatedAt, "PPPppp"),
        "Updated at": format(file.UpdatedAt, "PPPppp"),
    }
}