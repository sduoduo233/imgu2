# database

## users

| Name | Type | Description |
|---|---|---|
| id | INTEGER |  |
| username | TEXT | unique username |
| email | TEXT | unique email in lower case |
| password | TEXT | bcrypt hashed password |
| role | INTEGER | 0=admin 1=user 2=banned user |
| user_group | INTEGER | the group id which the user belongs to |
| user_group_expire | INTEGER | timestamp when the membership of the group expires | 

## social_logins

This table records social accounts linked to a user account.

| Name | Type | Description |
|---|---|---|
| id | INTEGER | |
| type | TEXT | `google` or `github` |
| user | INTEGER | user id |
| account_id | TEXT | the account id from oauth providers (e.g. google id or github id)

## storages

| Name | Type | Description |
|---|---|---|
| id | INTEGER | |
| name | TEXT | display name of this storage driver |
| type | TEXT | `local` / `s3` / `ftp` / `webdav` / `telegraph` |
| config | TEXT | JSON configuration for the driver |
| enabled | BOOLEAN | whether reading and writing is enabled for the driver |
| allow_upload | BOOLEAN | whether writing is enabled |

## images

| Name | Type | Description |
|---|---|---|
| id | INTEGER | |
| storage | INTEGER | storage id |
| uploader | INTEGER | user id (null represents guest user)
| file_name | TEXT | the display name for the file |
| internal_name | TEXT | the file name used in the corresponding storage driver |
| uploader_ip | TEXT | |
| time | INTEGER | timestamp when the image is uploaded |
| expire_time | INTEGER | timestamp when the image should be deleted |

## settings

key-value storage for settings

| Name | Type | Description |
|---|---|---|
| id | INTEGER | |
| key | TEXT | |
| value | TEXT | |

## sessions

| Name | Type | Description |
|---|---|---|
| id | INTEGER | |
| token | TEXT | token that should be presented in http cookies |
| user | INTEGER | user id |
| expire_at | INTEGER | timestamp when the session should no longer be valid |

## groups

| Name | Type | Description |
|---|---|---|
| id | INTEGER | |
| name | TEXT | display name of the group |
| allow_upload | BOOLEAN | whether users in the group are allowed to upload |
| max_file_size | INTEGER | maximum file size in bytes |
| upload_per_* | INTEGER | unused (not implemented) |
| total_uploads | INTEGER | unused (not implemented) |
| max_retention_seconds | INTEGER | The number of seconds an uploaded image is kept for before it is deleted. Zero means uploaded images are stored without a time limit. |

