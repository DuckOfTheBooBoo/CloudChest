export default interface Folder {
    ID: number;
    ParentID: number | null;
    UserID: number;
    Code: string;
    Name: string;
    HasChild: boolean;
    IsFavorite: boolean;
    CreatedAt: Date;
    UpdatedAt: Date;
    DeletedAt: Date | null;
    folderCode: string | null;
    ParentFolder: Folder | null;
}