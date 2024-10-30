import type Folder from "../models/folder";

export default function(obj: any): obj is Folder {
    return "HasChild" in obj;
}