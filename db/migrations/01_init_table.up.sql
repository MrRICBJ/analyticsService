create table if not exists analytics (
   id bigserial primary key,
   time timestamptz(0) not null,
   user_id uuid not null,
   data jsonb not null
);

create table if not exists raw_analytics (
   id bigserial primary key,
   time timestamptz(0) not null,
   user_id uuid not null,
   data jsonb not null,
   processed boolean default false
);
