create table chat
(
chat_id serial not null
constraint chat_pk
primary key,
course_id integer default 0 not null,
section_id integer default 0 not null,
level_id integer default 0 not null,
lesson_id integer default 0 not null,
student_id integer default 0 not null,
rating varchar(256) default ''::character varying not null,
ahtung boolean default false not null,
rating_teacher integer default 0 not null,
ahtung_teacher integer default 0 not null
);

alter table chat owner to school_user;

create unique index chat_chat_id_uindex
on chat (chat_id);








create table courses
(
	id serial not null
		constraint courses_pk
			primary key,
	name varchar(256) not null,
	cost real not null,
	created_at timestamp default now() not null,
	updated_at timestamp default now() not null,
	users integer default 0 not null,
	dz integer default 0 not null,
	sale integer default 0 not null,
	total integer default 0 not null,
	total_price_for_user integer default 0 not null
);

alter table courses owner to school_user;

create unique index courses_id_uindex
	on courses (id);








create table lesson_carousel
(
	course_id integer default 0 not null,
	section_id integer default 0 not null,
	level_id integer default 0 not null,
	lesson_array integer[]
);

alter table lesson_carousel owner to school_user;

create unique index lesson_carousel_level_id_uindex
	on lesson_carousel (level_id);






create table lessons
(
	course_id integer not null,
	section_id integer not null,
	level_id integer not null,
	name varchar(256) not null,
	description varchar(256),
	thesis text[],
	task varchar(256),
	created_at timestamp default now() not null,
	updated_at timestamp default now() not null,
	lesson_id serial not null
		constraint lessons_pk
			primary key,
	status_free boolean default false not null
);

alter table lessons owner to school_user;





create table levels
(
	course_id integer not null,
	section_id integer not null,
	level_id serial not null,
	name varchar(256) not null,
	created_at timestamp default now() not null,
	updated_at timestamp default now() not null
);

alter table levels owner to school_user;

create unique index levels_level_id_uindex
	on levels (level_id);






create table messages
(
	chat_id integer default 0 not null,
	message_id serial not null
		constraint messages_pk
			primary key,
	role varchar(256) default 'student'::character varying not null,
	text varchar(256) default 'none'::character varying not null,
	first_name varchar(256) default 'testDefaultName'::character varying not null,
	time_mes timestamp with time zone default now() not null,
	not_read boolean default true not null
);

alter table messages owner to school_user;

create unique index messages_message_id_uindex
	on messages (message_id);






create table recovery_pass
(
	email varchar(256) default 'nothing'::character varying not null,
	code varchar(256) not null
);

alter table recovery_pass owner to school_user;

create unique index recovery_pass_kode_uindex
	on recovery_pass (code);








create table request_log
(
	id serial not null
		constraint request_log_pk
			primary key,
	user_id integer default 0 not null,
	request_url varchar(255) default ''::character varying not null,
	created_at timestamp with time zone default now() not null
);

alter table request_log owner to school_user;

create unique index request_log_id_uindex
	on request_log (id);










create table section_and_teacher
(
	teacher_id integer default 0 not null,
	course_id integer default 0 not null,
	section_id integer default 0 not null
);

alter table section_and_teacher owner to school_user;







create table sections
(
	id serial not null
		constraint sections_pk
			primary key,
	course_id integer not null,
	name varchar(256) not null,
	created_at timestamp default now() not null,
	updated_at timestamp default now() not null
);

alter table sections owner to school_user;

create unique index sections_id_uindex
	on sections (id);






create table teacher_info
(
	id integer not null,
	good integer default 0 not null,
	improve integer default 0 not null,
	ahtung integer default 0 not null,
	answer_time_sec bigint default 0 not null,
	answer_count integer default 0 not null
);

alter table teacher_info owner to school_user;






create table users
(
	id serial not null
		constraint data_users_pk
			primary key,
	created_at timestamp default now(),
	updated_at timestamp default now() not null,
	email varchar(256) not null,
	pass varchar(256) not null,
	first_name varchar(256) not null,
	user_role varchar(256) not null,
	times_seconds bigint default 0 not null
);

alter table users owner to school_user;

create unique index data_users_email_uindex
	on users (email);

create unique index data_users_id_uindex
	on users (id);










