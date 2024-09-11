package models

import (
    "os"
    "time"

    "github.com/minio/minio-go/v7"
)

type MinioFileInfo struct {
    objInfo minio.ObjectInfo
}

// Implementing the FileInfo interface methods
func (m MinioFileInfo) Name() string {
    return m.objInfo.Key
}

func (m MinioFileInfo) Size() int64 {
    return m.objInfo.Size
}

func (m MinioFileInfo) Mode() os.FileMode {
    // MinIO doesn't have file permissions, so we'll return a default file mode
    return 0644
}

func (m MinioFileInfo) ModTime() time.Time {
    return m.objInfo.LastModified
}

func (m MinioFileInfo) IsDir() bool {
    return false // MinIO object is a file, not a directory
}

func (m MinioFileInfo) Sys() interface{} {
    return nil // No underlying system data
}