# web-gallery-app

## API DOCUMENTATION
### Files
#### Get file list

- Endpoint: `/api/files`
- Method: `GET`
- Headers:
    - Authorization: `Bearer (JWT)`
- Query parameters:
    - isTrashCan: `boolean (default: false)`
    - isFavorite: `boolean (default: false)`
    > isTrashCan and isFavorite cannot have the same value.
