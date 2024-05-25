import { MinIOFile } from "../models/file";
import { format } from "date-fns";

export function fileDetailFormatter(file: MinIOFile): Object {
  return {
    ID: file.ID,
    "File name": file.FileName,
    "File type": file.FileType,
    "File size": humanFileSize(file.FileSize, true),
    Location: file.StoragePath,
    "Created at": format(file.CreatedAt, "PPPppp"),
    "Updated at": format(file.UpdatedAt, "PPPppp"),
  };
}

/**
 * Format bytes as human-readable text.
 * @author https://stackoverflow.com/a/14919494
 * @param bytes Number of bytes.
 * @param si True to use metric (SI) units, aka powers of 1000. False to use
 *           binary (IEC), aka powers of 1024.
 * @param dp Number of decimal places to display.
 *
 * @return Formatted string.
 */
export function humanFileSize(
  bytes: number,
  si: boolean = false,
  dp: number = 1
): string {
  const thresh = si ? 1000 : 1024;

  if (Math.abs(bytes) < thresh) {
    return bytes + " B";
  }

  const units = si
    ? ["kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"]
    : ["KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"];
  let u = -1;
  const r = 10 ** dp;

  do {
    bytes /= thresh;
    ++u;
  } while (
    Math.round(Math.abs(bytes) * r) / r >= thresh &&
    u < units.length - 1
  );

  return bytes.toFixed(dp) + " " + units[u];
}
