export default interface Folder {
    ID: number;
    ParentID: number | null;
    UserID: number;
    Code: string;
    Name: string;
    CreatedAt: Date;
    UpdatedAt: Date;
    DeletedAt: Date | null;
}