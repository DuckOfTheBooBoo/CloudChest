Project CloudChest {
  database_type: 'MySQL'
}

Table users {
  id int [pk, increment, note: "generated by gorm"]
  first_name varchar(50) [not null]
  last_name varchar(50) [not null]
  email varchar(255) [unique, not null]
  password varchar(64) [not null]
  minio_bucket varchar(50) [not null]
  created_at datetime [not null]
  updated_at datetime
}

Table folders {
  id int [pk, increment, note: "generated by gorm"]
  user_id int [ref: > users.id]
  parent_id int [ref: > folders.id]
  code varchar(100) [not null]
  name varchar(255) [not null]
}

Table files {
  id int [pk, increment, note: "generated by gorm"]
  user_id int [ref: > users.id]
  folder_id int [ref: > folders.id]
  file_name varchar(255) [not null]
  file_size int [not null]
  file_type varchar(100) [not null]
  is_favorite bool [not null, default: 0]
}
