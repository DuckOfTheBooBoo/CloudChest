export interface FilePatchRequest {
    file_name?: string;
    is_favorite?: boolean;
    is_restore?: boolean;
    folder_code?: string;
}

export interface FolderPatchRequest {
    folder_name?: string;
    is_favorite?: boolean;
    is_restore?: boolean;
    parent_folder_code?: string;
}