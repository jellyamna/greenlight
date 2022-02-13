CREATE TABLE IF NOT EXISTS perusahaan (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    address text NOT NULL,
    tlp text  NULL,
    npwp text  NULL,
    rek text  NULL,
    ket text  NULL,
    version integer NOT NULL DEFAULT 1
);