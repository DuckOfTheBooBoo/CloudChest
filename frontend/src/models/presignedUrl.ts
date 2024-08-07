export interface PresignedURL {
    Scheme: string;
    Opaque: string;
    User: null;
    Host: string;
    Path: string;
    RawPath: string;
    OmitHost: boolean;
    ForceQuery: boolean;
    RawQuery: string;
    Fragment: string;
    RawFragment: string;
}
