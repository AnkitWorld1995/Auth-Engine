drop table if exists hvoc.users;
CREATE TABLE hvoc.users (
                            id uuid NOT NULL DEFAULT hvoc.uuid_generate_v4(),
                            user_id SERIAL,
                            user_name varchar NOT NULL,
                            first_name varchar NOT NULL,
                            last_name varchar NOT NULL,
                            "password" varchar NULL,
                            email varchar NOT NULL,
                            phone int4 NOT NULL,
                            address varchar NULL,
                            is_admin bool NULL DEFAULT false,
                            user_type varchar null,
                            created_at timestamptz NULL,
                            updated_at timestamptz NULL,
                            CONSTRAINT users_pkey PRIMARY KEY (user_id)
);


drop table if exists hvoc.upload_files;
create table hvoc.upload_files (
                                   id          uuid    default hvoc.uuid_generate_v4() primary key ,
                                   user_id     serial,
                                   upload_file jsonb,
                                   created_at  timestamptz,
                                   updated_at  timestamptz,
                                   constraint fk_user_id foreign key(user_id) references hvoc.users(user_id)
);