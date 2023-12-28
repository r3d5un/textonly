CREATE TABLE public.users (
	user_id int8 NOT NULL GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 9223372036854775807 START 1 CACHE 1 NO CYCLE),
	"name" varchar(100) NOT NULL,
	summary text NOT NULL,
	"content" text NOT NULL,
	CONSTRAINT users_pkey PRIMARY KEY (user_id)
);

CREATE TABLE public.socials (
	id bigserial NOT NULL,
	user_id int8 NOT NULL,
	social_platform varchar(100) NOT NULL,
	link varchar(1000) NOT NULL
);

ALTER TABLE public.socials ADD CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES public.users(user_id);

CREATE TABLE public.posts (
	id bigserial NOT NULL,
	title varchar(255) NOT NULL,
	lead text NOT NULL,
	post text NOT NULL,
	last_update timestamp NOT NULL DEFAULT NOW(),
	created timestamp NOT NULL DEFAULT NOW(),
	CONSTRAINT posts_pkey PRIMARY KEY (id)
);
