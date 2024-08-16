import type {CancelTokenSource} from "axios";

export default interface RequestData {
    id: string;
    filename: string;
    request: Promise<any>;
    cancelToken: CancelTokenSource;
    progress: number;
}
