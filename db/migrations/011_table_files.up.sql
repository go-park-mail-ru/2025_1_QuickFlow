create table if not exists files(
                                    id int generated always as identity primary key,
                                    file_url text not null,
                                    filename text not null,
                                    created_at timestamptz not null default now(),
                                    unique(file_url, filename)
);