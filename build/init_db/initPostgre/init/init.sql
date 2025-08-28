CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    fio TEXT,
    username TEXT UNIQUE,
    pass TEXT,
    role TEXT,
    age SMALLINT,
    specialty TEXT,
    price INTEGER,
    rating REAL DEFAULT 0,
    teachers UUID[],
    students UUID[],
    requests UUID[]
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    teacher_id UUID NOT NULL,
    student_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    student_fio VARCHAR(255) NOT NULL,
    teacher_fio VARCHAR(255) NOT NULL,
    file_name_task VARCHAR(255),
    file_data_task BYTEA,
    file_name_solution VARCHAR(255),
    file_data_solution BYTEA,
    grade INTEGER,
    status VARCHAR(50) NOT NULL,
    FOREIGN KEY (teacher_id) REFERENCES users(id),
    FOREIGN KEY (student_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS chat_rooms (
  id        UUID PRIMARY KEY,
  user1_id  UUID NOT NULL,
  user2_id  UUID NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE IF NOT EXISTS chat_messages (
  id        UUID PRIMARY KEY,
  room_id   UUID REFERENCES chat_rooms(id) ON DELETE CASCADE,
  sender_id UUID NOT NULL,
  text      TEXT NOT NULL,
  sent_at   TIMESTAMPTZ NOT NULL,
  status    SMALLINT NOT NULL -- 1-sent, 2-delivered, 3-read
);
