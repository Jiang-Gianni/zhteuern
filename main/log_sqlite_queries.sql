-- name: WriteLog :exec
insert into log_table(
    msg, attributes, log_level, time_string, time_unix,
    app, env, commit_git, src_file, src_line, src_function
)
values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
