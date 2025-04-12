# 🌥️ CloudChest

CloudChest is a "Google Drive"-like web application focused on core file and folder management functionalities. It is designed as a personal project to showcase back-end and full-stack development skills, with features such as media previewing, user authentication, soft delete support, and more.

> ⚠️ **Note:** CloudChest is not intended for production use. This project serves as a demonstration of core skills and system architecture.

---

## 🚀 Features

### ✅ Completed

- 🔐 **User Authentication & Authorization**
  - Secure login and registration
  - Role-based access control (basic)

- 📂 **File Management**
  - Upload, rename, delete files
  - Soft delete, permanent delete, and restoration

- 🗂️ **Folder Management**
  - Create, rename, delete, and move folders
  - Hierarchical folder structure

- 🎞️ **Media Preview**
  - Image thumbnails
  - Video streaming (HLS)

### 🔜 Planned / In Progress

- ✏️ Text editor for `.txt`, `.md`, or `.json` files  
- 📄 PDF reader  
- 🧠 Bug tracking and cleanup  

### ❌ Not Yet Implemented (Known Gaps)

- 🌍 Multi-user object access controls (public/private links)
- 📦 User storage limits and tiered access
- 👥 User management dashboard (admin tools)

---

## 🛠️ Tech Stack

- **Frontend:** Vue.js
- **Backend:** Go (with Gin-gonic framework)
- **Database:** MariaDB (RDBMS) / MinIO (object storage)
- **Media:** FFmpeg (for video streaming)
- **Auth:** JWT-based
- **Containerization:** Docker

---

## 📦 Deployment (Docker)

This app is intended to be deployed locally using Docker. Below is a simplified step-by-step guide.

### Prerequisites
- Docker and Docker Compose installed

### Steps

1. Clone this repository:
   ```bash
   git clone https://github.com/DuckOfTheBooBoo/CloudChest.git
   cd CloudChest
   ```

2. Modify the environment variables below in `docker-compose.yml` with your preferred values.
    ```yaml
    minio:
        environment:
            MINIO_ROOT_PASSWORD: password # Change this to a secure password

    db:
        environment:
            MARIADB_ROOT_PASSWORD: password # Change this to a secure password
    ```

3. Build and run using Docker Compose:
    ```bash
    docker compose up -d --build
    ```

4. Access the app at `http://localhost:8080`.

## ⚠ Known Issues

### Grandparent folder is not restored
When a file within a folder hierarchy (e.g., file <- parent folder <- grandparent folder) is deleted, followed by deleting its parent and grandparent folders, attempting to restore the file from the trash will not restore the grandparent folder. This happens because recursive checks to restore parent folders only trigger if the immediate parent folder is also deleted.

```
file <- parent folder <- grandparent folder
```
IF we delete `file`, and later delete `grandparent folder`
```
file (deleted) <- parent folder <- grandparent folder (deleted)
```
Attempting to restore `file (deleted)` **WILL NOT** restore `grandparent folder (deleted)`, thus ending up with this state
```
file <- parent folder <- grandparent folder (deleted)
```
This will make the `parent folder` and `file` not showing up in the file explorer. The current workaround is to also restore `grandparent folder (deleted)`

> Fixes are planned in the future.

### Image thumbnail doesn't show up once it finished uploading
This is because the thumbnail is generated after the upload is successful in a goroutine.

> TODO: implement a polling mechanism to check if the thumbnail is generated and show it once it's ready.