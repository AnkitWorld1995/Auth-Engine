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
CREATE TABLE hvoc.upload_files (
                                   id uuid NOT NULL DEFAULT hvoc.uuid_generate_v4(),
                                   user_id serial4 NOT NULL,
                                   product_id int4,
                                   catalog_id int4,
                                   upload_file jsonb NULL,
                                   created_at timestamptz NULL,
                                   updated_at timestamptz NULL,
                                   CONSTRAINT upload_files_pkey PRIMARY KEY (id),
                                   CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES hvoc.users(user_id),
                                   constraint product_fkey foreign key (product_id) references  hvoc.products (product_id),
                                   constraint catalog_fkey foreign key (catalog_id) references  hvoc.catalog (catalog_id)
);



drop table if exist hvoc.products;
CREATE TABLE hvoc.products (
                               id uuid NOT NULL DEFAULT hvoc.uuid_generate_v4(),
                               product_id SERIAL,
                               manufacturer varchar NOT NULL,
                               brand varchar NOT NULL,
                               qty int4 NOT NULL,
                               price float8 default 0,
                               product_data jsonb NULL ,
                               created_at timestamptz NULL,
                               updated_at timestamptz NULL,
                               CONSTRAINT product_pkey PRIMARY KEY (product_id)
);

drop table if exists hvoc.catalog;
CREATE TABLE hvoc.catalog (
                              id uuid NOT NULL DEFAULT hvoc.uuid_generate_v4(),
                              catalog_id SERIAL,
                              product_id int4,
                              created_at timestamptz NULL,
                              updated_at timestamptz NULL,
                              CONSTRAINT catalog_pkey PRIMARY KEY (catalog_id),
                              constraint product_fkey foreign key (product_id) references  hvoc.products (product_id)
);


CREATE TABLE hvoc.uploaded_items (
                                     id uuid NOT NULL DEFAULT hvoc.uuid_generate_v4(),
                                     product_id SERIAL,
                                     manufacturer varchar NOT NULL,
                                     brand varchar NOT NULL,
                                     qty int4 NOT NULL,
                                     price float8 default 0,
                                     product_data jsonb NULL ,
                                     created_at timestamptz NULL,
                                     updated_at timestamptz NULL,
                                     CONSTRAINT product_pkey PRIMARY KEY (product_id)
);