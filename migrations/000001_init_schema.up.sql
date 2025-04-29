-- ================================================
-- Таблица пользователей
-- ================================================
CREATE TABLE users
(
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    full_name     VARCHAR(255)        NOT NULL,
    is_teacher    BOOLEAN             NOT NULL DEFAULT FALSE
);

-- ================================================
-- Таблица учебных заведений
-- ================================================
CREATE TABLE institutions
(
    id   SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- ================================================
-- Таблица групп (создается ДО студентов и учителей)
-- ================================================
CREATE TABLE groups
(
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(255) NOT NULL,
    institution_id INT          NOT NULL,
    FOREIGN KEY (institution_id) REFERENCES institutions (id)
);

-- ================================================
-- Таблицы студентов и преподавателей
-- ================================================
CREATE TABLE teachers
(
    user_id        INT PRIMARY KEY,
    institution_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (institution_id) REFERENCES institutions (id)
);

CREATE TABLE students
(
    user_id        INT PRIMARY KEY,
    institution_id INT NOT NULL,
    group_id       INT,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (institution_id) REFERENCES institutions (id),
    FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE SET NULL
);

-- ================================================
-- Таблица предметов
-- ================================================
CREATE TABLE subjects
(
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(255) NOT NULL,
    institution_id INT          NOT NULL,
    FOREIGN KEY (institution_id) REFERENCES institutions (id)
);

-- ================================================
-- Новая таблица инициалов преподавателей
-- ================================================
CREATE TABLE teacher_initials
(
    id             SERIAL PRIMARY KEY,
    initials       VARCHAR(50) NOT NULL,
    teacher_id     INT,
    institution_id INT         NOT NULL,
    FOREIGN KEY (teacher_id) REFERENCES teachers (user_id) ON DELETE SET NULL,
    FOREIGN KEY (institution_id) REFERENCES institutions (id) ON DELETE CASCADE,
    UNIQUE (initials, institution_id)
);

-- ================================================
-- Таблица расписания
-- ================================================
CREATE TABLE schedule
(
    id                  SERIAL PRIMARY KEY,
    group_id            INT         NOT NULL,
    subject_id          INT         NOT NULL,
    date                DATE        NOT NULL,
    pair_number         INT         NOT NULL,
    classroom           VARCHAR(50) NOT NULL,
    teacher_initials_id INT,
    start_time          TIME        NOT NULL,
    end_time            TIME        NOT NULL,
    FOREIGN KEY (group_id) REFERENCES groups (id),
    FOREIGN KEY (subject_id) REFERENCES subjects (id),
    FOREIGN KEY (teacher_initials_id) REFERENCES teacher_initials (id)
);


-- ================================================
-- Таблица чатов
-- ================================================
CREATE TABLE chats
(
    id          SERIAL PRIMARY KEY,
    group_id    INT                 NOT NULL,
    owner_id    INT                 NOT NULL,
    subject_id  INT                 NOT NULL,
    join_code   VARCHAR(20) UNIQUE  NOT NULL,
    invite_link VARCHAR(255) UNIQUE NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups (id),
    FOREIGN KEY (subject_id) REFERENCES subjects (id),
    FOREIGN KEY (owner_id) REFERENCES teachers (user_id)
);

-- ================================================
-- Таблицы сообщений и файлов
-- ================================================

CREATE TABLE messages
(
    id                SERIAL PRIMARY KEY,
    chat_id           INT NOT NULL,
    user_id           INT NOT NULL,
    text              TEXT,
    message_group_id  INT,
    parent_message_id INT,
    created_at        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (message_group_id) REFERENCES messages (id),
    FOREIGN KEY (parent_message_id) REFERENCES messages (id) ON DELETE SET NULL
);


CREATE TABLE message_files
(
    id         SERIAL PRIMARY KEY,
    message_id INT          NOT NULL,
    file_url   VARCHAR(255) NOT NULL,
    FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE
);

-- ================================================
-- Таблица для связи студентов и чатов (многие ко многим)
-- ================================================
CREATE TABLE student_chats
(
    student_id INT NOT NULL,
    chat_id    INT NOT NULL,
    PRIMARY KEY (student_id, chat_id),
    FOREIGN KEY (student_id) REFERENCES students (user_id) ON DELETE CASCADE,
    FOREIGN KEY (chat_id) REFERENCES chats (id) ON DELETE CASCADE
);

-- ================================================
-- Таблица избранных сообщений
-- ================================================
CREATE TABLE favorites
(
    user_id    INT NOT NULL,
    message_id INT NOT NULL,
    PRIMARY KEY (user_id, message_id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (message_id) REFERENCES messages (id)
);

-- ================================================
-- Таблицы опросов
-- ================================================
CREATE TABLE polls
(
    id         SERIAL PRIMARY KEY,
    chat_id    INT  NOT NULL,
    question   TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats (id)
);

CREATE TABLE poll_options
(
    id          SERIAL PRIMARY KEY,
    poll_id     INT  NOT NULL,
    option_text TEXT NOT NULL,
    FOREIGN KEY (poll_id) REFERENCES polls (id)
);

CREATE TABLE votes
(
    user_id        INT NOT NULL,
    poll_option_id INT NOT NULL,
    PRIMARY KEY (user_id, poll_option_id),
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (poll_option_id) REFERENCES poll_options (id)
);

-- ================================================
-- Таблица почтовых масок учебных заведений
-- ================================================
CREATE TABLE institution_email_masks
(
    id             SERIAL PRIMARY KEY,
    institution_id INT                 NOT NULL,
    email_mask     VARCHAR(255) UNIQUE NOT NULL,
    FOREIGN KEY (institution_id) REFERENCES institutions (id) ON DELETE CASCADE
);

-- ================================================
-- Уникальные индексы
-- ================================================
CREATE UNIQUE INDEX subjects_unique ON subjects (name, institution_id);
CREATE UNIQUE INDEX schedule_unique ON schedule (group_id, date, pair_number);
CREATE UNIQUE INDEX chats_unique ON chats (group_id, subject_id, owner_id);
CREATE UNIQUE INDEX groups_unique ON groups (name, institution_id);

