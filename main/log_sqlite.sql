create table if not exists log_table(
    msg text not null default '',
    attributes text not null default '',
    log_level int not null default 0,
    time_string text not null default '',
    time_unix int not null default 0,
    app text not null default '',
    env text not null default '',
    commit_git text not null default '',
    src_file text not null default '',
    src_line int not null default 0,
    src_function text not null default ''
);
